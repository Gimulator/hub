package client

import (
	"context"
	"sync"

	hubv1 "github.com/Gimulator/hub/api/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Client is a wrapper for controller-runtime client
type Client struct {
	client.Client

	Scheme *runtime.Scheme

	roomM sync.Mutex
}

// NewClient returns new instance of Client
func NewClient(c client.Client, scheme *runtime.Scheme) (Client, error) {
	return Client{
		Client: c,
		Scheme: scheme,
	}, nil
}

////////////
/// Room ///
////////////

// SyncRoom takes a Room object and updates it or creates it if not exists
func (c *Client) SyncRoom(ctx context.Context, room *hubv1.Room) (*hubv1.Room, error) {
	c.roomM.Lock()
	defer c.roomM.Unlock()

	syncedRoom := &hubv1.Room{
		ObjectMeta: metav1.ObjectMeta{
			Name:      room.Name,
			Namespace: room.Namespace,
		},
	}

	err := retry.RetryOnConflict(retry.DefaultBackoff,
		func() error {
			_, err := controllerutil.CreateOrUpdate(ctx, c.Client, syncedRoom, func() error {
				syncedRoom.Status = *room.Status.DeepCopy()
				return nil
			})
			return err
		},
	)

	return syncedRoom, err
}

// GetRoom gets a namespaced Room object from Kubernetes
func (c *Client) GetRoom(ctx context.Context, nn types.NamespacedName) (*hubv1.Room, error) {
	room := &hubv1.Room{}

	return room, c.Get(ctx, nn, room)
}

// DeleteRoom deletes a Room object from Kubernetes
func (c *Client) DeleteRoom(ctx context.Context, room *hubv1.Room) error {
	if err := c.Delete(ctx, room); !errors.IsNotFound(err) {
		return err
	}
	return nil
}
