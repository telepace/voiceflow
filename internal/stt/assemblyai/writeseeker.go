package assemblyai

import (
	"bytes"
	"fmt"
	"io"
)

type BufferWriteSeeker struct {
	buf bytes.Buffer
	pos int64
}

func (bws *BufferWriteSeeker) Write(p []byte) (n int, err error) {
	if bws.pos != int64(bws.buf.Len()) {
		// 如果当前位置不在缓冲区末尾，需要调整缓冲区
		data := bws.buf.Bytes()
		if bws.pos < int64(len(data)) {
			bws.buf.Truncate(int(bws.pos))
		}
	}
	n, err = bws.buf.Write(p)
	bws.pos += int64(n)
	return n, err
}

func (bws *BufferWriteSeeker) Seek(offset int64, whence int) (int64, error) {
	var newPos int64
	switch whence {
	case io.SeekStart:
		newPos = offset
	case io.SeekCurrent:
		newPos = bws.pos + offset
	case io.SeekEnd:
		newPos = int64(bws.buf.Len()) + offset
	default:
		return 0, fmt.Errorf("invalid whence")
	}
	if newPos < 0 {
		return 0, fmt.Errorf("negative position")
	}
	bws.pos = newPos
	return bws.pos, nil
}

func (bws *BufferWriteSeeker) Bytes() []byte {
	return bws.buf.Bytes()
}
