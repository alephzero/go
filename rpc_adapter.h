#ifndef A0_GO_RPC_ADAPTER_H
#define A0_GO_RPC_ADAPTER_H

#include <a0/rpc.h>

#include "alloc_adapter.h"
#include "packet_adapter.h"

extern void a0go_rpc_request_callback(void*, a0_rpc_request_t);

A0_STATIC_INLINE
a0_err_t a0go_rpc_server_init(a0_rpc_server_t* server,
                              a0_rpc_topic_t topic,
                              uintptr_t alloc_id,
                              uintptr_t onrequest_id,
                              uintptr_t oncancel_id) {
  a0_alloc_t alloc = {
      .user_data = (void*)alloc_id,
      .alloc = a0go_alloc,
      .dealloc = NULL,
  };
  a0_rpc_request_callback_t onrequest = {
      .user_data = (void*)onrequest_id,
      .fn = a0go_rpc_request_callback,
  };
  a0_packet_id_callback_t oncancel = {
      .user_data = (void*)oncancel_id,
      .fn = a0go_packet_id_callback,
  };
  return a0_rpc_server_init(server, topic, alloc, onrequest, oncancel);
}

A0_STATIC_INLINE
a0_err_t a0go_rpc_client_init(a0_rpc_client_t* client,
                              a0_rpc_topic_t topic,
                              uintptr_t alloc_id) {
  a0_alloc_t alloc = {
      .user_data = (void*)alloc_id,
      .alloc = a0go_alloc,
      .dealloc = NULL,
  };
  return a0_rpc_client_init(client, topic, alloc);
}

A0_STATIC_INLINE
a0_err_t a0go_rpc_send(a0_rpc_client_t* client,
                       a0_packet_t pkt,
                       uintptr_t packet_callback_id) {
  a0_packet_callback_t packet_callback = {
      .user_data = (void*)packet_callback_id,
      .fn = a0go_packet_callback,
  };
  return a0_rpc_client_send(client, pkt, packet_callback);
}

#endif  // A0_GO_RPC_ADAPTER_H
