package client

import (
	"context"
	"sync"

	hubv1 "github.com/Gimulator/hub/api/v1"
	corev1 "k8s.io/api/core/v1"
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
func NewClient(c client.Client, scheme *runtime.Scheme) (*Client, error) {
	return &Client{
		Client: c,
		Scheme: scheme,
	}, nil
}

///////////////////////////////////////////////////
////////////////////////////////////////// Room ///
///////////////////////////////////////////////////

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

// GetRoom takes a NamespacedName key and returns a Room object if exists
func (c *Client) GetRoom(ctx context.Context, key types.NamespacedName) (*hubv1.Room, error) {
	room := &hubv1.Room{}

	return room, c.Get(ctx, key, room)
}

// DeleteRoom deletes a Room object
func (c *Client) DeleteRoom(ctx context.Context, room *hubv1.Room) error {
	if err := c.Delete(ctx, room); !errors.IsNotFound(err) {
		return err
	}
	return nil
}

//////////////////////////////////////////////////
////////////////////////////////////////// Pod ///
//////////////////////////////////////////////////

// SyncPod takes a Pod object and updates it or creates it if not exists
func (c *Client) SyncPod(ctx context.Context, pod *corev1.Pod, owner metav1.Object) (*corev1.Pod, error) {
	syncedPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		},
	}

	err := retry.RetryOnConflict(retry.DefaultBackoff,
		func() error {
			_, err := controllerutil.CreateOrUpdate(ctx, c.Client, syncedPod, func() error {
				syncedPod.Annotations = pod.Annotations
				syncedPod.Labels = pod.Labels
				syncedPod.Spec = pod.Spec
				controllerutil.SetOwnerReference(owner, syncedPod, c.Scheme)
				return nil
			})
			return err
		},
	)

	return syncedPod, err
}

// GetPod takes a NamespacedName key and returns a Pod object if exists
func (c *Client) GetPod(ctx context.Context, key types.NamespacedName) (*corev1.Pod, error) {
	pod := &corev1.Pod{}

	return pod, c.Get(ctx, key, pod)
}

// DeletePod deletes a Pod object
func (c *Client) DeletePod(ctx context.Context, pod *corev1.Pod) error {
	if err := c.Delete(ctx, pod); !errors.IsNotFound(err) {
		return err
	}
	return nil
}

//////////////////////////////////////////////////
////////////////////////////////////////// PVC ///
//////////////////////////////////////////////////

// SyncPVC takes a PVC object and updates it or creates it if not exists
func (c *Client) SyncPVC(ctx context.Context, pvc *corev1.PersistentVolumeClaim, owner metav1.Object) (*corev1.PersistentVolumeClaim, error) {
	syncedPVC := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvc.Name,
			Namespace: pvc.Namespace,
		},
	}

	err := retry.RetryOnConflict(retry.DefaultBackoff,
		func() error {
			_, err := controllerutil.CreateOrUpdate(ctx, c.Client, syncedPVC, func() error {
				syncedPVC.Spec = *pvc.Spec.DeepCopy()
				controllerutil.SetOwnerReference(owner, syncedPVC, c.Scheme)
				return nil
			})
			return err
		},
	)

	return syncedPVC, err
}

// GetPVC takes a NamespacedName key and returns a PVC object if exists
func (c *Client) GetPVC(ctx context.Context, key types.NamespacedName) (*corev1.PersistentVolumeClaim, error) {
	pvc := &corev1.PersistentVolumeClaim{}

	return pvc, c.Get(ctx, key, pvc)
}

// DeletePVC deletes a PVC object
func (c *Client) DeletePVC(ctx context.Context, pvc *corev1.PersistentVolumeClaim) error {
	if err := c.Delete(ctx, pvc); !errors.IsNotFound(err) {
		return err
	}
	return nil
}

//////////////////////////////////////////////////
////////////////////////////////////// Service ///
//////////////////////////////////////////////////

// SyncService takes a Service object and updates it or creates it if not exists
func (c *Client) SyncService(ctx context.Context, service *corev1.Service, owner metav1.Object) (*corev1.Service, error) {
	syncedService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      service.Name,
			Namespace: service.Namespace,
		},
	}

	err := retry.RetryOnConflict(retry.DefaultBackoff,
		func() error {
			_, err := controllerutil.CreateOrUpdate(ctx, c.Client, syncedService, func() error {
				syncedService.Spec = *service.Spec.DeepCopy()
				controllerutil.SetOwnerReference(owner, syncedService, c.Scheme)
				return nil
			})
			return err
		},
	)

	return syncedService, err
}

//////////////////////////////////////////////////
///////////////////////////////////// ConfigMap///
//////////////////////////////////////////////////

func (c *Client) SyncConfigMap(ctx context.Context, cm *corev1.ConfigMap, owner metav1.Object) (*corev1.ConfigMap, error) {
	syncedConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cm.Name,
			Namespace: cm.Namespace,
		},
	}

	err := retry.RetryOnConflict(retry.DefaultBackoff,
		func() error {
			_, err := controllerutil.CreateOrUpdate(ctx, c.Client, syncedConfigMap, func() error {
				newcm := cm.DeepCopy()
				syncedConfigMap.Data = newcm.Data
				syncedConfigMap.Annotations = newcm.Annotations
				syncedConfigMap.Labels = newcm.Labels
				if owner != nil {
					controllerutil.SetOwnerReference(owner, syncedConfigMap, c.Scheme)
				}
				return nil
			})
			return err
		},
	)

	return syncedConfigMap, err
}

func (c *Client) GetConfigMap(ctx context.Context, key types.NamespacedName) (*corev1.ConfigMap, error) {
	cm := &corev1.ConfigMap{}

	return cm, c.Get(ctx, key, cm)
}
