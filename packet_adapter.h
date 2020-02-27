#ifndef A0_GO_PACKET_ADAPTER_H
#define A0_GO_PACKET_ADAPTER_H

#include <a0/packet.h>

#include "common_adapter.h"

extern void a0go_packet_callback(void*, a0_packet_t);
extern void a0go_packet_id_callback(void*, char*);

#endif  // A0_GO_PACKET_ADAPTER_H
