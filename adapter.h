#ifndef A0_GO_ADAPTER_H
#define A0_GO_ADAPTER_H

#include <a0/packet.h>
// #include <a0/pubsub.h>

extern void a0go_alloc(void*, size_t, a0_buf_t*);
// extern void a0go_callback(void*);
// extern void a0go_subscriber_callback(void*, a0_packet_t);

static inline errno_t a0go_packet_build(size_t num_headers,
                                        a0_packet_header_t* headers,
                                        a0_buf_t payload,
                                        int alloc_id,
                                        a0_packet_t* out) {
  a0_alloc_t alloc = {
      .user_data = &alloc_id,
      .fn = a0go_alloc,
  };
  return a0_packet_build(num_headers, headers, payload, alloc, out);
}

// static inline errno_t a0go_subscriber_sync_next(a0_subscriber_sync_t* sub_sync,
//                                                 a0_packet_t* pkt,
//                                                 void* user_data) {
//   a0_alloc_t alloc = {
//       .user_data = user_data,
//       .fn = a0go_alloc,
//   };
//   return a0_subscriber_sync_next(sub_sync, alloc, pkt);
// }

// static inline errno_t a0go_subscriber_init_unmapped(a0_subscriber_t* sub,
//                                                     const char* container,
//                                                     const char* topic,
//                                                     a0_subscriber_read_start_t read_start,
//                                                     a0_subscriber_read_next_t read_next,
//                                                     void* user_data) {
//   a0_alloc_t alloc = {
//       .user_data = user_data,
//       .fn = a0go_alloc,
//   };
//   a0_subscriber_callback_t callback = {
//       .user_data = user_data,
//       .fn = a0go_subscriber_callback,
//   };
//   return a0_subscriber_init_unmapped(sub, container, topic, read_start, read_next, alloc, callback);
// }

// static inline errno_t a0go_subscriber_close(a0_subscriber_t* sub, void* user_data) {
//   a0_callback_t callback = {
//       .user_data = user_data,
//       .fn = a0go_callback,
//   };
//   return a0_subscriber_close(sub, callback);
// }

#endif  // A0_GO_ADAPTER_H
