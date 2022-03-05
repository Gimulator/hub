package reporter

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	hubv1 "github.com/Gimulator/hub/api/v1"
	"github.com/Gimulator/hub/pkg/client"
	"github.com/Gimulator/hub/pkg/mq"
	"github.com/Gimulator/hub/pkg/name"
	"github.com/Gimulator/hub/pkg/s3"
	"github.com/Gimulator/protobuf/go/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

type Reporter struct {
	token        string
	rabbit       *mq.Rabbit
	client       *client.Client
	k8sClientSet *kubernetes.Clientset
}

func NewReporter(token string, rabbit *mq.Rabbit, client *client.Client, k8sClientSet *kubernetes.Clientset) (*Reporter, error) {
	return &Reporter{
		token:        token,
		rabbit:       rabbit,
		client:       client,
		k8sClientSet: k8sClientSet,
	}, nil
}

func (r *Reporter) Report(ctx context.Context, room *hubv1.Room) (bool, error) {
	reports := r.prepareReports(room)

	switch room.Status.GimulatorStatus {
	case corev1.PodSucceeded:
		if err := r.informBackendS3(ctx, room); err != nil {
			return false, err
		}
		return true, nil
	case corev1.PodRunning:
		if !room.Spec.TerminateOnActorFailure {
			return false, r.informGimulator(ctx, room, reports)
		}

		shouldDelete, err := r.checkPodsForFailure(ctx, room)
		if err != nil {
			return true, err
		}
		return shouldDelete, r.informGimulator(ctx, room, reports)
	case corev1.PodFailed:
		result := &api.Result{
			Id:     room.Spec.ID,
			Status: api.Result_failed,
			Msg:    "Gimulaor failed",
		}
		// TODO: should write better result for backend
		if err := r.informRabbit(room, result); err != nil {
			return false, err
		}
		return true, nil
	default:
		// Gimulator is not still ready, We will inform it in the next call of reconciler
		return false, nil
	}
}

func (r *Reporter) ReportTimeout(room *hubv1.Room, threshold int64) error {
	result := &api.Result{
		Id:     room.Spec.ID,
		Status: api.Result_failed,
		Msg:    fmt.Sprintf("Timeout limit exceeded (%d seconds).", threshold),
	}
	// TODO: should write better result for backend
	return r.informRabbit(room, result)
}

func (r *Reporter) checkPodsForFailure(ctx context.Context, room *hubv1.Room) (bool, error) {
	for actor, status := range room.Status.ActorStatuses {
		if status == corev1.PodFailed {
			key := types.NamespacedName{Name: name.ActorPodName(actor), Namespace: room.Namespace}
			pod, err := r.client.GetPod(ctx, key)
			if err != nil {
				return true, err
			}

			var stream io.ReadCloser

			if err := r.GetPodLogs(ctx, r.k8sClientSet, pod, &corev1.PodLogOptions{
				Timestamps: true,
			}, &stream); err != nil {
				return true, err
			}

			log, err := io.ReadAll(stream)
			if err != nil {
				return true, err
			}

			result := &api.Result{
				Id:     room.Spec.ID,
				Status: api.Result_failed,
				Msg:    fmt.Sprintf("Actor faced an exception.\n%s", string(log)),
			}
			if err := r.informRabbit(room, result); err != nil {
				return true, err
			}
			return true, nil
		}
	}
	if status := room.Status.DirectorStatus; status == corev1.PodFailed {
		key := types.NamespacedName{Name: name.DirectorPodName(room.Spec.Director.Name), Namespace: room.Namespace}
		pod, err := r.client.GetPod(ctx, key)
		if err != nil {
			return true, err
		}

		var stream io.ReadCloser

		if err := r.GetPodLogs(ctx, r.k8sClientSet, pod, &corev1.PodLogOptions{
			Timestamps: true,
		}, &stream); err != nil {
			return true, err
		}

		log, err := io.ReadAll(stream)
		if err != nil {
			return true, err
		}

		result := &api.Result{
			Id:     room.Spec.ID,
			Status: api.Result_failed,
			Msg:    fmt.Sprintf("Director faced an exception.\n%s", string(log)),
		}
		if err := r.informRabbit(room, result); err != nil {
			return true, err
		}
		return true, nil
	}
	return false, nil
}

func (r *Reporter) prepareReports(room *hubv1.Room) []*api.Report {
	reports := make([]*api.Report, 0)

	reports = append(reports, &api.Report{
		Name:   room.Spec.Director.Name,
		Status: r.kubeToAPIStatus(room.Status.DirectorStatus),
	})

	for name, phase := range room.Status.ActorStatuses {
		reports = append(reports, &api.Report{
			Name:   name,
			Status: r.kubeToAPIStatus(phase),
		})
	}

	return reports
}

func (r *Reporter) kubeToAPIStatus(phase corev1.PodPhase) api.Status {
	status := api.Status_unknown
	switch phase {
	case corev1.PodRunning:
		status = api.Status_running
	case corev1.PodFailed:
		status = api.Status_failed
	case corev1.PodSucceeded:
		status = api.Status_succeeded
	}
	return status
}

func (r *Reporter) informGimulator(ctx context.Context, room *hubv1.Room, reports []*api.Report) error {
	address := name.GimulatorServiceName(room.Spec.ID) + ":" + strconv.Itoa(name.GimulatorServicePort())

	ctx2, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx2, address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return fmt.Errorf("could not connect to Gimulator with address=%v", address)
	}
	defer conn.Close()

	client := api.NewOperatorAPIClient(conn)
	for _, report := range reports {
		ctx := metadata.AppendToOutgoingContext(ctx, "token", r.token)
		if _, err := client.SetUserStatus(ctx, report); err != nil {
			return err
		}
	}
	return nil
}

func (r *Reporter) informRabbit(room *hubv1.Room, result *api.Result) error {
	if err := r.rabbit.Send(result); err != nil {
		return err
	}
	return nil
}

func (r *Reporter) informBackendS3(ctx context.Context, room *hubv1.Room) error {
	// Setting up K8s CLientSet
	podLogOpts := &corev1.PodLogOptions{
		Timestamps: true,
	}

	var stream io.ReadCloser

	// Dumping logs
	// Actor(s)
	for _, actor := range room.Spec.Actors {
		key := types.NamespacedName{Name: name.ActorPodName(actor.Name), Namespace: room.Namespace}
		actorPod, err := r.client.GetPod(ctx, key)
		if err != nil {
			return err
		}

		if err := r.GetPodLogs(ctx, r.k8sClientSet, actorPod, podLogOpts, &stream); err != nil {
			return err
		}

		if err := s3.PutObject(ctx, stream, name.S3LogsBucket(), name.S3LogObjectNameForActor(room.Spec.ID, actor.Name)); err != nil {
			return err
		}
	}

	// Director
	directorKey := types.NamespacedName{Name: name.DirectorPodName(room.Spec.Director.Name), Namespace: room.Namespace}
	directorPod, err := r.client.GetPod(ctx, directorKey)
	if err != nil {
		return err
	}

	if err := r.GetPodLogs(ctx, r.k8sClientSet, directorPod, podLogOpts, &stream); err != nil {
		return err
	}

	if err := s3.PutObject(ctx, stream, name.S3LogsBucket(), name.S3LogObjectNameForDirector(room.Spec.ID, room.Spec.Director.Name)); err != nil {
		return err
	}

	// Gimulator
	// TODO: There's a freakin bug lying below. For some reason, gimulator logs can't make it to the S3. Yeah you might not need Gimulator's logs but still ... why is this happening?

	// gimulatorKey := types.NamespacedName{Name: name.GimulatorPodName(room.Spec.ID), Namespace: room.Namespace}
	// gimulatorPod, err := r.client.GetPod(ctx, gimulatorKey)
	// if err != nil {
	// 	return err
	// }

	// if err := r.GetPodLogs(ctx, clientSet, gimulatorPod, podLogOpts, &stream); err != nil {
	// 	return err
	// }

	// if err := s3.PutObject(ctx, stream, name.S3LogsBucket(), name.S3LogObjectName(room.Spec.ID, "gimulator")); err != nil {
	// 	return err
	// }

	return nil
}

func (r *Reporter) GetPodLogs(ctx context.Context, clientSet *kubernetes.Clientset, pod *corev1.Pod, options *corev1.PodLogOptions, reader *io.ReadCloser) error {
	req := clientSet.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, options)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return err
	}
	*reader = podLogs
	return nil
}
