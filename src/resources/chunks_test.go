package resources

import (
	"bytes"
	"io"

	"depozit-server/src/config"
	"depozit-server/src/security"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Frame size of a full chunk
const record = chunkHeaderSize + config.MaxEncryptedChunkSize

func newCipher(t *testing.T) *security.Cipher {
	c, err := security.NewCipher(make([]byte, 32))
	require.NoError(t, err)

	return c
}

func encrypt(t *testing.T, c *security.Cipher, data []byte, path string) []byte {
	var out bytes.Buffer

	w := newChunkWriter(&out, c, path)

	// Odd write sizes, so chunk boundaries do not match write boundaries
	for len(data) > 0 {
		n := min(7777, len(data))
		_, err := w.Write(data[:n])
		require.NoError(t, err)
		data = data[n:]
	}

	require.NoError(t, w.Close())

	return out.Bytes()
}

func decrypt(c *security.Cipher, enc []byte, path string) ([]byte, error) {
	r := newChunkReader(bytes.NewReader(enc), c, path)
	out := []byte{}

	for {
		p, err := r.Next()

		if err == io.EOF {
			return out, nil
		}

		if err != nil {
			return nil, err
		}

		out = append(out, p...)
	}
}

func data(size int) []byte {
	d := make([]byte, size)

	for i := range d {
		d[i] = byte(i * 31)
	}

	return d
}

func TestRoundTrip(t *testing.T) {
	c := newCipher(t)

	for _, size := range []int{0, 1, config.ChunkSize - 1, config.ChunkSize, config.ChunkSize + 1, 3 * config.ChunkSize, 3*config.ChunkSize + 5} {
		enc := encrypt(t, c, data(size), "a.txt")

		got, err := decrypt(c, enc, "a.txt")
		require.NoError(t, err, size)
		assert.Equal(t, data(size), got, size)
	}
}

func TestChunks(t *testing.T) {
	c := newCipher(t)
	enc := encrypt(t, c, data(2*config.ChunkSize+10), "a.txt")

	r := newChunkReader(bytes.NewReader(enc), c, "a.txt")

	var sizes []int
	var got []byte

	for {
		p, err := r.Next()

		if err == io.EOF {
			break
		}

		require.NoError(t, err)
		sizes = append(sizes, len(p))
		got = append(got, p...)
	}

	assert.Equal(t, []int{config.ChunkSize, config.ChunkSize, 10}, sizes)
	assert.Equal(t, data(2*config.ChunkSize+10), got)
}

// Small writes are stored in full chunks, so disk size does not depend on how the client splits data
func TestSmallWritesAreRechunked(t *testing.T) {
	c := newCipher(t)

	var out bytes.Buffer
	w := newChunkWriter(&out, c, "a.txt")

	for range 1000 {
		_, err := w.Write([]byte("x"))
		require.NoError(t, err)
	}

	require.NoError(t, w.Close())

	assert.Equal(t, chunkHeaderSize+1000+config.MaxEncryptedChunkSize-config.ChunkSize, out.Len())
}

func TestWrongPath(t *testing.T) {
	c := newCipher(t)
	enc := encrypt(t, c, data(10), "a.txt")

	_, err := decrypt(c, enc, "b.txt")
	assert.ErrorIs(t, err, errCorrupt)
}

func TestWrongKey(t *testing.T) {
	enc := encrypt(t, newCipher(t), data(10), "a.txt")

	other, err := security.NewCipher(bytes.Repeat([]byte{1}, 32))
	require.NoError(t, err)

	_, err = decrypt(other, enc, "a.txt")
	assert.ErrorIs(t, err, errCorrupt)
}

func TestTampered(t *testing.T) {
	c := newCipher(t)
	enc := encrypt(t, c, data(3*config.ChunkSize+10), "a.txt")

	chunk := func(b []byte, i int) []byte {
		start := i * record
		return b[start:min(start+record, len(b))]
	}

	join := func(parts ...[]byte) []byte { return bytes.Join(parts, nil) }

	cases := map[string][]byte{
		"bit flip":              append(append([]byte{}, enc[:len(enc)-1]...), enc[len(enc)-1]^1),
		"truncated at boundary": enc[:2*record],
		"truncated mid chunk":   enc[:2*record+5],
		"empty":                 {},
		"chunks swapped":        join(chunk(enc, 1), chunk(enc, 0), chunk(enc, 2), chunk(enc, 3)),
		"chunk duplicated":      join(enc, chunk(enc, 0)),
	}

	for name, tampered := range cases {
		_, err := decrypt(c, tampered, "a.txt")
		assert.ErrorIs(t, err, errCorrupt, name)
	}
}
