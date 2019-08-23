package alephzero

/*
#cgo pkg-config: alephzero
#include <a0/topic_manager.h>
#include <stdlib.h>  // free
*/
import "C"

import (
	"unsafe"
)

type TopicManager struct {
	c C.a0_topic_manager_t
}

func NewTopicManagerFromJSON(json string) (tm TopicManager, err error) {
	cJson := C.CString(json)
	defer C.free(unsafe.Pointer(cJson))

	err = errorFrom(C.a0_topic_manager_init_jsonstr(&tm.c, cJson))
	return
}

func (tm *TopicManager) Close() error {
	return errorFrom(C.a0_topic_manager_close(&tm.c))
}

func (tm *TopicManager) ConfigTopic() (shm ShmObj, err error) {
	err = errorFrom(C.a0_topic_manager_config_topic(&tm.c, &shm.c))
	return
}

func (tm *TopicManager) PublisherTopic(name string) (shm ShmObj, err error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	err = errorFrom(C.a0_topic_manager_publisher_topic(&tm.c, cName, &shm.c))
	return
}

func (tm *TopicManager) SubscriberTopic(name string) (shm ShmObj, err error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	err = errorFrom(C.a0_topic_manager_subscriber_topic(&tm.c, cName, &shm.c))
	return
}

func (tm *TopicManager) RpcServerTopic(name string) (shm ShmObj, err error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	err = errorFrom(C.a0_topic_manager_rpc_server_topic(&tm.c, cName, &shm.c))
	return
}

func (tm *TopicManager) RpcClientTopic(name string) (shm ShmObj, err error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	err = errorFrom(C.a0_topic_manager_rpc_client_topic(&tm.c, cName, &shm.c))
	return
}

func (tm *TopicManager) Unref(shm ShmObj) error {
	return errorFrom(C.a0_topic_manager_unref(&tm.c, shm.c))
}
