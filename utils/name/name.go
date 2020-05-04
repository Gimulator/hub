package name

import "fmt"

func BucketDashKey(bucket, key string) string {
	return fmt.Sprintf("%s-%s", bucket, key)
}

func NameDashID(name string, id int) string {
	return fmt.Sprintf("%s-%d", name, id)
}

func GimulatorConfigMap(id int) string {
	return fmt.Sprintf("roles-%d", id)
}

func TerminatedFile(name string) string {
	return fmt.Sprintf("%s-terminated", name)
}
