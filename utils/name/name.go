package name

import "fmt"

func ConfigMapName(bucket, key string) string {
	return fmt.Sprintf("%s-%s", bucket, key)
}

func ContainerName(name string, id int) string {
	return fmt.Sprintf("%s-%d", name, id)
}

func RoomJobName(id int) string {
	return fmt.Sprintf("%d", id)
}

func MLJobName(id int) string {
	return fmt.Sprintf("%d", id)
}

func TerminatedFileName(name string) string {
	return fmt.Sprintf("%s-terminated", name)
}
