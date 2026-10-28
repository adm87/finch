package tmx

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/adm87/finch/pkg/errs"
)

// GID is a global tile ID. The top bits say how the tile is flipped.
type GID uint32

const (
	gidFlipH     GID = 0x80000000
	gidFlipV     GID = 0x40000000
	gidFlipD     GID = 0x20000000
	gidRotateHex GID = 0x10000000
	gidFlags         = gidFlipH | gidFlipV | gidFlipD | gidRotateHex
)

// ID is the GID without its flip bits. 0 means no tile.
func (g GID) ID() uint32 {
	return uint32(g &^ gidFlags)
}

func (g GID) FlippedHorizontally() bool {
	return g&gidFlipH != 0
}

func (g GID) FlippedVertically() bool {
	return g&gidFlipV != 0
}

// FlippedDiagonally swaps x and y. On hexagonal maps it rotates 60 degrees.
func (g GID) FlippedDiagonally() bool {
	return g&gidFlipD != 0
}

// RotatedHex120 only applies to hexagonal maps.
func (g GID) RotatedHex120() bool {
	return g&gidRotateHex != 0
}

type DataEncoding string

const (
	// EncodingNone stores each tile as a <tile> element.
	EncodingNone   DataEncoding = ""
	EncodingCSV    DataEncoding = "csv"
	EncodingBase64 DataEncoding = "base64"
)

func (e DataEncoding) IsValid() bool {
	switch e {
	case EncodingNone, EncodingCSV, EncodingBase64:
		return true
	default:
		return false
	}
}

type DataCompression string

const (
	CompressionNone DataCompression = ""
	CompressionGzip DataCompression = "gzip"
	CompressionZlib DataCompression = "zlib"
	CompressionZstd DataCompression = "zstd"
)

func (c DataCompression) IsValid() bool {
	switch c {
	case CompressionNone, CompressionGzip, CompressionZlib, CompressionZstd:
		return true
	default:
		return false
	}
}

func validateEncoding(enc DataEncoding, comp DataCompression) error {
	var errList []error
	if !enc.IsValid() {
		errList = append(errList, errs.Argument("encoding", fmt.Sprintf("unknown encoding %q", enc)))
	}
	if !comp.IsValid() {
		errList = append(errList, errs.Argument("compression", fmt.Sprintf("unknown compression %q", comp)))
	}
	if comp != CompressionNone && enc != EncodingBase64 {
		errList = append(errList, errs.Argument("compression", "only works with base64 encoding"))
	}
	return errors.Join(errList...)
}

type DataTile struct {
	GID GID `xml:"gid,attr"`
}

// Chunk is a piece of an infinite map's layer. It's encoded the same way
// as the layer data that holds it.
type Chunk struct {
	Data   string     `xml:",chardata"`
	Tiles  []DataTile `xml:"tile"`
	X      int        `xml:"x,attr"`
	Y      int        `xml:"y,attr"`
	Width  int        `xml:"width,attr"`
	Height int        `xml:"height,attr"`
}

type LayerData struct {
	Data        string          `xml:",chardata"`
	Encoding    DataEncoding    `xml:"encoding,attr"`
	Compression DataCompression `xml:"compression,attr"`
	Tiles       []DataTile      `xml:"tile"`
	Chunks      []Chunk         `xml:"chunk"`
}

// GIDs decodes a finite layer's tiles, row by row.
func (d LayerData) GIDs() ([]GID, error) {
	return decodeGIDs(d.Data, d.Tiles, d.Encoding, d.Compression)
}

// ChunkGIDs decodes one of the layer's chunks, row by row.
func (d LayerData) ChunkGIDs(chunk Chunk) ([]GID, error) {
	return decodeGIDs(chunk.Data, chunk.Tiles, d.Encoding, d.Compression)
}

func (d LayerData) Validate() error {
	errList := []error{validateEncoding(d.Encoding, d.Compression)}
	for i, c := range d.Chunks {
		if c.Width <= 0 {
			errList = append(errList, prefixed(fmt.Sprintf("chunk[%d]", i), errs.ArgumentOutOfRange("width", c.Width, errMustBePositive)))
		}
		if c.Height <= 0 {
			errList = append(errList, prefixed(fmt.Sprintf("chunk[%d]", i), errs.ArgumentOutOfRange("height", c.Height, errMustBePositive)))
		}
	}
	return errors.Join(errList...)
}

// validateLayout checks the data is stored the way the map expects. Infinite
// maps use chunks and finite maps don't.
func (d LayerData) validateLayout(infinite bool) error {
	hasData := strings.TrimSpace(d.Data) != "" || len(d.Tiles) > 0
	switch {
	case infinite && hasData:
		return errs.Argument("data", "infinite maps store tiles in chunks")
	case !infinite && len(d.Chunks) > 0:
		return errs.Argument("data", "only infinite maps use chunks")
	case !infinite && !hasData:
		return errs.Argument("data", errMustNotBeEmpty)
	}
	return nil
}

func decodeGIDs(raw string, tiles []DataTile, enc DataEncoding, comp DataCompression) ([]GID, error) {
	switch enc {
	case EncodingNone:
		gids := make([]GID, len(tiles))
		for i, t := range tiles {
			gids[i] = t.GID
		}
		return gids, nil
	case EncodingCSV:
		return decodeCSV(raw)
	case EncodingBase64:
		b, err := decodeBase64(raw, comp)
		if err != nil {
			return nil, err
		}
		if len(b)%4 != 0 {
			return nil, errs.Format("data", fmt.Sprintf("%d bytes is not a whole number of tiles", len(b)))
		}
		gids := make([]GID, len(b)/4)
		for i := range gids {
			gids[i] = GID(binary.LittleEndian.Uint32(b[i*4:]))
		}
		return gids, nil
	default:
		return nil, errs.Argument("encoding", fmt.Sprintf("unknown encoding %q", enc))
	}
}

func decodeCSV(raw string) ([]GID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	fields := strings.Split(raw, ",")
	gids := make([]GID, len(fields))
	for i, f := range fields {
		v, err := strconv.ParseUint(strings.TrimSpace(f), 10, 32)
		if err != nil {
			return nil, errs.Format("data", fmt.Sprintf("tile %d", i)).Wrap(err)
		}
		gids[i] = GID(v)
	}
	return gids, nil
}

func decodeBase64(raw string, comp DataCompression) ([]byte, error) {
	b, err := base64.StdEncoding.DecodeString(strings.TrimSpace(raw))
	if err != nil {
		return nil, errs.Format("data", "invalid base64").Wrap(err)
	}

	var r io.ReadCloser
	switch comp {
	case CompressionNone:
		return b, nil
	case CompressionGzip:
		r, err = gzip.NewReader(bytes.NewReader(b))
	case CompressionZlib:
		r, err = zlib.NewReader(bytes.NewReader(b))
	case CompressionZstd:
		return nil, errs.NotSupported("zstd compression")
	default:
		return nil, errs.Argument("compression", fmt.Sprintf("unknown compression %q", comp))
	}
	if err != nil {
		return nil, errs.Format("data", "invalid "+string(comp)).Wrap(err)
	}
	defer func() { _ = r.Close() }()

	out, err := io.ReadAll(r)
	if err != nil {
		return nil, errs.Format("data", "invalid "+string(comp)).Wrap(err)
	}
	return out, nil
}
