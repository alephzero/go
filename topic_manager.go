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

type TopicAliasTarget struct {
	Container string `json:"container"`
	Topic     string `json:"topic"`
}

type TopicManager struct {
	Container         string                      `json:"container"`
	SubscriberAliases map[string]TopicAliasTarget `json:"subscriber_aliases,omitempty"`
	RpcClientAliases  map[string]TopicAliasTarget `json:"rpc_client_aliases,omitempty"`
	PrpcClientAliases map[string]TopicAliasTarget `json:"prpc_client_aliases,omitempty"`
}

func (tm *TopicManager) withC(fn func(C.a0_topic_manager_t)) {
	ctm := C.a0_topic_manager_t{}
	ctm.container = C.CString(tm.Container)
	defer C.free(unsafe.Pointer(ctm.container))

	ctm.subscriber_aliases_size = C.uint64_t(len(tm.SubscriberAliases))
	if ctm.subscriber_aliases_size == 0 {
		ctm.subscriber_aliases = nil
	} else {
		ctm.subscriber_aliases = (*C.a0_topic_alias_t)(C.malloc(ctm.subscriber_aliases_size * C.size_t(unsafe.Sizeof(C.a0_topic_alias_t{}))))
		defer C.free(unsafe.Pointer(ctm.subscriber_aliases))

		cSubAliases := (*[1 << 30]C.a0_topic_alias_t)(unsafe.Pointer(ctm.subscriber_aliases))[:int(ctm.subscriber_aliases_size):int(ctm.subscriber_aliases_size)]

		i := 0
		for k, v := range tm.SubscriberAliases {
			cSubAliases[i].name = C.CString(k)
			cSubAliases[i].target_container = C.CString(v.Container)
			cSubAliases[i].target_topic = C.CString(v.Topic)

			defer C.free(unsafe.Pointer(cSubAliases[i].name))
			defer C.free(unsafe.Pointer(cSubAliases[i].target_container))
			defer C.free(unsafe.Pointer(cSubAliases[i].target_topic))

			i++
		}
	}

	ctm.rpc_client_aliases_size = C.uint64_t(len(tm.RpcClientAliases))
	if ctm.rpc_client_aliases_size == 0 {
		ctm.rpc_client_aliases = nil
	} else {
		ctm.rpc_client_aliases = (*C.a0_topic_alias_t)(C.malloc(ctm.rpc_client_aliases_size * C.size_t(unsafe.Sizeof(C.a0_topic_alias_t{}))))
		defer C.free(unsafe.Pointer(ctm.rpc_client_aliases))

		cRpcAliases := (*[1 << 30]C.a0_topic_alias_t)(unsafe.Pointer(ctm.rpc_client_aliases))[:int(ctm.rpc_client_aliases_size):int(ctm.rpc_client_aliases_size)]

		i := 0
		for k, v := range tm.RpcClientAliases {
			cRpcAliases[i].name = C.CString(k)
			cRpcAliases[i].target_container = C.CString(v.Container)
			cRpcAliases[i].target_topic = C.CString(v.Topic)

			defer C.free(unsafe.Pointer(cRpcAliases[i].name))
			defer C.free(unsafe.Pointer(cRpcAliases[i].target_container))
			defer C.free(unsafe.Pointer(cRpcAliases[i].target_topic))

			i++
		}
	}

	ctm.prpc_client_aliases_size = C.uint64_t(len(tm.PrpcClientAliases))
	if ctm.prpc_client_aliases_size == 0 {
		ctm.prpc_client_aliases = nil
	} else {
		ctm.prpc_client_aliases = (*C.a0_topic_alias_t)(C.malloc(ctm.prpc_client_aliases_size * C.size_t(unsafe.Sizeof(C.a0_topic_alias_t{}))))
		defer C.free(unsafe.Pointer(ctm.prpc_client_aliases))

		cPrpcAliases := (*[1 << 30]C.a0_topic_alias_t)(unsafe.Pointer(ctm.prpc_client_aliases))[:int(ctm.prpc_client_aliases_size):int(ctm.prpc_client_aliases_size)]

		i := 0
		for k, v := range tm.PrpcClientAliases {
			cPrpcAliases[i].name = C.CString(k)
			cPrpcAliases[i].target_container = C.CString(v.Container)
			cPrpcAliases[i].target_topic = C.CString(v.Topic)

			defer C.free(unsafe.Pointer(cPrpcAliases[i].name))
			defer C.free(unsafe.Pointer(cPrpcAliases[i].target_container))
			defer C.free(unsafe.Pointer(cPrpcAliases[i].target_topic))

			i++
		}
	}

	fn(ctm)
}

func (tm *TopicManager) OpenConfigTopic() (shm Shm, err error) {
	tm.withC(func(ctm C.a0_topic_manager_t) {
		err = errorFrom(C.a0_topic_manager_open_config_topic(&ctm, &shm.c))
	})
	return
}

func (tm *TopicManager) OpenPublisherTopic(name string) (shm Shm, err error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	tm.withC(func(ctm C.a0_topic_manager_t) {
		err = errorFrom(C.a0_topic_manager_open_publisher_topic(&ctm, cName, &shm.c))
	})
	return
}

func (tm *TopicManager) OpenSubscriberTopic(name string) (shm Shm, err error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	tm.withC(func(ctm C.a0_topic_manager_t) {
		err = errorFrom(C.a0_topic_manager_open_subscriber_topic(&ctm, cName, &shm.c))
	})
	return
}

func (tm *TopicManager) OpenRpcServerTopic(name string) (shm Shm, err error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	tm.withC(func(ctm C.a0_topic_manager_t) {
		err = errorFrom(C.a0_topic_manager_open_rpc_server_topic(&ctm, cName, &shm.c))
	})
	return
}

func (tm *TopicManager) OpenRpcClientTopic(name string) (shm Shm, err error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	tm.withC(func(ctm C.a0_topic_manager_t) {
		err = errorFrom(C.a0_topic_manager_open_rpc_client_topic(&ctm, cName, &shm.c))
	})
	return
}

func (tm *TopicManager) OpenPrpcServerTopic(name string) (shm Shm, err error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	tm.withC(func(ctm C.a0_topic_manager_t) {
		err = errorFrom(C.a0_topic_manager_open_prpc_server_topic(&ctm, cName, &shm.c))
	})
	return
}

func (tm *TopicManager) OpenPrpcClientTopic(name string) (shm Shm, err error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	tm.withC(func(ctm C.a0_topic_manager_t) {
		err = errorFrom(C.a0_topic_manager_open_prpc_client_topic(&ctm, cName, &shm.c))
	})
	return
}
