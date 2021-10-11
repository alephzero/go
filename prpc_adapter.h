#ifndef A0_GO_PRPC_ADAPTER_H
#define A0_GO_PRPC_ADAPTER_H

#include <a0/prpc.h>

#include "alloc_adapter.h"
#include "packet_adapter.h"

extern void a0go_prpc_connection_callback(void*, a0_prpc_connection_t);

A0_STATIC_INLINE
a0_err_t a0go_prpc_server_init(a0_prpc_server_t* server,
                               a0_prpc_topic_t topic,
                               uintptr_t alloc_id,
                               uintptr_t onconnect_id,
                               uintptr_t oncancel_id) {
  a0_alloc_t alloc = {
      .user_data = (void*)alloc_id,
      .alloc = a0go_alloc,
      .dealloc = NULL,
  };
  a0_prpc_connection_callback_t onconnection = {
      .user_data = (void*)onconnect_id,
      .fn = a0go_prpc_connection_callback,
  };
  a0_packet_id_callback_t oncancel = {
      .user_data = (void*)oncancel_id,
      .fn = a0go_packet_id_callback,
  };
  return a0_prpc_server_init(server, topic, alloc, onconnection, oncancel);
}

A0_STATIC_INLINE
a0_err_t a0go_prpc_client_init(a0_prpc_client_t* client,
                               a0_prpc_topic_t topic,
                               uintptr_t alloc_id) {
  a0_alloc_t alloc = {
      .user_data = (void*)alloc_id,
      .alloc = a0go_alloc,
      .dealloc = NULL,
  };
  return a0_prpc_client_init(client, topic, alloc);
}

extern void a0go_prpc_progress_callback(void*, a0_packet_t, bool);

A0_STATIC_INLINE
a0_err_t a0go_prpc_client_connect(a0_prpc_client_t* client,
                                  a0_packet_t pkt,
                                  uintptr_t prpc_progress_callback_id) {
  a0_prpc_progress_callback_t progress_callback = {
      .user_data = (void*)prpc_progress_callback_id,
      .fn = a0go_prpc_progress_callback,
  };
  return a0_prpc_client_connect(client, pkt, progress_callback);
}

#endif  // A0_GO_PRPC_ADAPTER_H
