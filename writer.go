package alephzero

/*
#cgo pkg-config: alephzero
#include <a0/writer.h>
*/
import "C"

type Writer struct {
	c C.a0_writer_t
}

func NewWriter(arena Arena) (w *Writer, err error) {
	w = &Writer{}
	err = errorFrom(C.a0_writer_init(&w.c, arena.c))
	return
}

func (w *Writer) Close() error {
	return errorFrom(C.a0_writer_close(&w.c))
}

func (w *Writer) Write(pkt Packet) error {
	cPkt := pkt.c()
	defer freeCPacket(cPkt)
	return errorFrom(C.a0_writer_write(&w.c, cPkt))
}

func (w *Writer) Push(m Middleware) error {
	return errorFrom(C.a0_writer_push(&w.c, m.c))
}

func (w *Writer) Wrap(m Middleware) (w2 Writer, err error) {
	err = errorFrom(C.a0_writer_wrap(&w.c, m.c, &w2.c))
	return
}
