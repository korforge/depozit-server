package resources

import (
	"depozit-server/src/config"
	"depozit-server/src/security"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"slices"
)

// Stored chunk: [length 4B][last 1B][nonce, ciphertext, tag]. The additional data is path || uint64(index) || last,
// so a chunk only opens at its own position in its own file, and a file only reads to the end when its last chunk is there
const chunkHeaderSize = 5

// Wraps every error that means the stored data is wrong, other errors come from the disk
var errCorrupt = errors.New("file is corrupted")

var (
	errChunkTooLarge = fmt.Errorf("%w: chunk too large", errCorrupt)
	errBadLastFlag   = fmt.Errorf("%w: invalid last chunk flag", errCorrupt)
	errTruncated     = fmt.Errorf("%w: file ends before its last chunk", errCorrupt)
	errTrailingData  = fmt.Errorf("%w: data after the last chunk", errCorrupt)
)

// Builds the additional data buffer for path. Each chunk only rewrites the 9 bytes at the end
func newChunkAD(path string) []byte {
	return append([]byte(path), make([]byte, 9)...)
}

func setChunkAD(ad []byte, index uint64, last bool) {
	binary.BigEndian.PutUint64(ad[len(ad)-9:], index)
	ad[len(ad)-1] = 0

	if last {
		ad[len(ad)-1] = 1
	}
}

func writeChunk(w io.Writer, cipher *security.Cipher, buf []byte, data []byte, ad []byte, last bool) ([]byte, error) {
	buf = cipher.Encrypt(append(buf[:0], 0, 0, 0, 0, 0), data, ad)
	binary.BigEndian.PutUint32(buf, uint32(len(buf)-chunkHeaderSize))

	if last {
		buf[4] = 1
	}

	_, err := w.Write(buf)

	return buf, err
}

// Seals data in fixed ChunkSize chunks, so the stored chunks do not depend on the client message size
type chunkWriter struct {
	w       io.Writer
	cipher  *security.Cipher
	ad      []byte
	index   uint64
	pending []byte
	buf     []byte
}

func newChunkWriter(w io.Writer, cipher *security.Cipher, path string) *chunkWriter {
	return &chunkWriter{w: w, cipher: cipher, ad: newChunkAD(path), pending: make([]byte, 0, config.ChunkSize)}
}

func (c *chunkWriter) seal(data []byte, last bool) error {
	setChunkAD(c.ad, c.index, last)
	c.index++

	var err error
	c.buf, err = writeChunk(c.w, c.cipher, c.buf, data, c.ad, last)

	return err
}

func (c *chunkWriter) Write(data []byte) (int, error) {
	n := len(data)

	for len(data) > 0 {
		// Full chunks with nothing pending are sealed straight from data, without a copy
		if len(c.pending) == 0 && len(data) >= config.ChunkSize {
			if err := c.seal(data[:config.ChunkSize], false); err != nil {
				return n - len(data), err
			}

			data = data[config.ChunkSize:]
			continue
		}

		k := min(len(data), config.ChunkSize-len(c.pending))
		c.pending = append(c.pending, data[:k]...)
		data = data[k:]

		if len(c.pending) == config.ChunkSize {
			if err := c.seal(c.pending, false); err != nil {
				return n - len(data), err
			}

			c.pending = c.pending[:0]
		}
	}

	return n, nil
}

// Writes the pending data as the last chunk, which can be empty. Call it once after the last Write
func (c *chunkWriter) Close() error {
	err := c.seal(c.pending, true)
	c.pending = c.pending[:0]

	return err
}

// Opens the chunks written by chunkWriter in order, and fails when a chunk is missing, moved, or edited
type chunkReader struct {
	r      io.Reader
	cipher *security.Cipher
	ad     []byte
	index  uint64
	done   bool
	enc    []byte
	data   []byte
}

func newChunkReader(r io.Reader, cipher *security.Cipher, path string) *chunkReader {
	return &chunkReader{r: r, cipher: cipher, ad: newChunkAD(path)}
}

// Returns the next decrypted chunk, valid until the next call. Returns io.EOF only after the last chunk
func (c *chunkReader) Next() ([]byte, error) {
	var header [chunkHeaderSize]byte

	_, err := io.ReadFull(c.r, header[:])

	if c.done {
		if err == io.EOF {
			return nil, io.EOF
		}

		if err == nil || err == io.ErrUnexpectedEOF {
			return nil, errTrailingData
		}

		return nil, err
	}

	if err == io.EOF || err == io.ErrUnexpectedEOF {
		return nil, errTruncated
	}

	if err != nil {
		return nil, err
	}

	n := binary.BigEndian.Uint32(header[:4])

	if n > config.MaxEncryptedChunkSize {
		return nil, errChunkTooLarge
	}

	if header[4] > 1 {
		return nil, errBadLastFlag
	}

	last := header[4] == 1
	c.enc = slices.Grow(c.enc[:0], int(n))[:n]

	if _, err := io.ReadFull(c.r, c.enc); err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil, fmt.Errorf("%w: %w", errCorrupt, io.ErrUnexpectedEOF)
		}

		return nil, err
	}

	setChunkAD(c.ad, c.index, last)

	if c.data, err = c.cipher.Decrypt(c.data[:0], c.enc, c.ad); err != nil {
		return nil, fmt.Errorf("%w: %w", errCorrupt, err)
	}

	c.index++
	c.done = last

	return c.data, nil
}
