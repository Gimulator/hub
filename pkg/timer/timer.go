package timer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-logr/logr"

	hubv1 "github.com/Gimulator/hub/api/v1"
	"github.com/Gimulator/hub/pkg/client"
	"github.com/Gimulator/hub/pkg/name"
	"github.com/Gimulator/hub/pkg/reporter"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Timer struct {
	clientSet *kubernetes.Clientset
	hubClient *client.Client
	timers    map[string]uint64 // map each pod to its timeout threshold
	mutex     sync.Mutex
	log       logr.Logger
	reporter  *reporter.Reporter
}

func NewTimer(clientSet *kubernetes.Clientset, log logr.Logger, reporter *reporter.Reporter, client *client.Client) (*Timer, error) {
	logger := log.WithValues("package", "Timer")

	return &Timer{
		clientSet: clientSet,
		hubClient: client,
		timers:    make(map[string]uint64),
		log:       logger,
		reporter:  reporter,
	}, nil
}

// SyncTimers
// StartTimer initiates a timer for every actor/director pod if necessary
// At the moment, every pod timer is supplied with the same threshold value but the code is ready to accept various values as threshold for any pod.
func (t *Timer) SyncTimers(room *hubv1.Room) {
	if room.Spec.Timeout <= 0 {
		return
	}

	//directorPodName := name.DirectorPodName(room.Spec.Director.Name)
	//if _, ok := t.timers[directorPodName]; ok {
	//	t.log.WithValues("podName", directorPodName).Info("Timer for pod exists.")
	//} else {
	//	t.timers[directorPodName] = room.Spec.Timeout
	//	go t.startPodTimer(directorPodName, room)
	//}

	for _, actor := range room.Spec.Actors {
		actorPodName := name.ActorPodName(actor.Name)
		if _, ok := t.timers[actorPodName]; ok {
			t.log.WithValues("podName", actorPodName).Info("Timer for pod exists.")
		} else {
			t.timers[actorPodName] = room.Spec.Timeout
			go t.startPodTimer(actorPodName, room)
		}
	}
}

// startPodTimer measures the age of a running actor/director pod and kills the room if a pod's age exceeds the given limit.
// Please note that this function DOES return an error object (if there's any) but because it's supposed to run as a goroutine you cannot actually receive/observe the given error.
// TODO: Error handling of this method needs some work.
// TODO: Will be more efficient to use time.NewTimer
func (t *Timer) startPodTimer(podName string, room *hubv1.Room) error {
	ctx := context.TODO()

	startTime, err := t.waitForPod(ctx, room, podName, -1) // TODO: not sure if notFoundThreshold should be dynamic or not.
	if err != nil {
		t.deleteTimer(podName)
		return err
	}

	for {
		isRunning, err := t.isPodRunning(ctx, room, podName)
		if err != nil {
			t.deleteTimer(podName)
			return err
		}

		if !isRunning {
			t.deleteTimer(podName)
			return nil
		}

		timeout, ok := t.timers[podName]

		// The timer has been deleted by another goroutine
		// FIXME: Use a better approach to avoid this
		if !ok {
			return nil
		}

		// Pod time has reached the timeout
		if diff := time.Now().Sub(startTime); diff.Seconds() > float64(timeout) {
			break
		}

		time.Sleep(time.Second * 2)
	}

	// limit has been exceeded. Must report and terminate the room.
	t.log.Info(fmt.Sprintf("Pod '%s' has reached the timeout threshold. Terminating the room.", podName))

	// Report result to rabbit
	if err := t.reporter.ReportTimeout(room, t.timers[podName]); err != nil {
		return err
	}

	// Kill all the timers for the room since it's timed out
	t.deleteTimers(room)

	// Delete the room
	return t.hubClient.DeleteRoom(ctx, room)
}

// waitForPod waits until the pod is in Running phase and then returns the time its status has changed to running state.
// `notFoundRetries` will be ignored if -1 is passed.
func (t *Timer) waitForPod(ctx context.Context, room *hubv1.Room, podName string, notFoundRetries int) (time.Time, error) {
	var pod *corev1.Pod
	var err error

	for {
		if pod, err = t.clientSet.CoreV1().Pods(room.Namespace).Get(ctx, podName, metav1.GetOptions{}); err != nil {
			// Return the error if retrieving the pod had any errors other than "Not Found"
			if !errors.IsNotFound(err) {
				return time.Time{}, err
			}

			// Pod is not found
			if notFoundRetries -= 1; notFoundRetries == -1 {
				return time.Time{}, fmt.Errorf("could not find pod '%s'", podName)
			}
			continue
		}

		// Pod has been found. Checking if its container is in running state
		for _, containerStatus := range pod.Status.ContainerStatuses {
			t.log.Info(fmt.Sprintf("Pod %s => %v", podName, pod.Status.ContainerStatuses))
			if containerStatus.State.Running != nil {
				return containerStatus.State.Running.StartedAt.Time, nil
			}
		}

		time.Sleep(time.Second * 1)
	}
}

// isPodRunning checks if the pod is in running state or not.
func (t *Timer) isPodRunning(ctx context.Context, room *hubv1.Room, podName string) (bool, error) {
	var pod *corev1.Pod
	var err error

	if pod, err = t.clientSet.CoreV1().Pods(room.Namespace).Get(ctx, podName, metav1.GetOptions{}); err != nil {
		return false, err
	}

	// Pod has been found. Checking if its container is in terminated state
	for _, containerStatus := range pod.Status.ContainerStatuses {
		t.log.Info(fmt.Sprintf("Pod %s => %v", podName, pod.Status.ContainerStatuses))
		if containerStatus.State.Terminated != nil {
			return false, nil
		}
	}

	return true, nil
}

// deleteTimers kills all the timers for a room.
func (t *Timer) deleteTimers(room *hubv1.Room) {
	//t.deleteTimer(name.DirectorPodName(room.Spec.Director.Name))
	for _, actor := range room.Spec.Actors {
		t.deleteTimer(name.ActorPodName(actor.Name))
	}
}

// deleteTimer kills the timer for a pod.
func (t *Timer) deleteTimer(podName string) {
	t.mutex.Lock()

	delete(t.timers, name.ActorPodName(podName))

	t.mutex.Unlock()
}
