package reporter

import (
	"context"
	"fmt"
	"strconv"
	"time"

	hubv1 "github.com/Gimulator/hub/api/v1"
	"github.com/Gimulator/hub/pkg/mq"
	"github.com/Gimulator/hub/pkg/name"
	"github.com/Gimulator/protobuf/go/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	corev1 "k8s.io/api/core/v1"
)

type Reporter struct {
	token  string
	rabbit *mq.Rabbit
}

func NewReporter(token string, rabbit *mq.Rabbit) (*Reporter, error) {
	return &Reporter{
		token:  token,
		rabbit: rabbit,
	}, nil
}

func (r *Reporter) Report(ctx context.Context, room *hubv1.Room) (bool, error) {
	reports := r.prepareReports(room)

	switch room.Status.GimulatorStatus {
	case corev1.PodSucceeded:
		return true, nil
	case corev1.PodRunning:
		return false, r.informGimulator(ctx, room, reports)
	case corev1.PodFailed:
		err := r.informBackend(room)
		if err != nil {
			return false, err
		}
		return true, nil
	default:
		// Gimulator is not still ready, We will inform it in the next call of reconciler
		return false, nil
	}
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

func (r *Reporter) informBackend(room *hubv1.Room) error {
	// TODO: should write better result for backend
	result := &api.Result{
		Id:     room.Spec.ID,
		Status: api.Result_failed,
		Msg:    "Gimulaor failed",
	}
	if err := r.rabbit.Send(result); err != nil {
		return err
	}
	return nil
}
