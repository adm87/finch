package tmx_test

import (
	"fmt"
	"testing"

	"github.com/go-openapi/testify/require"

	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/plugins/tiled/tmx"
)

// wantGIDs is the 2x2 pattern every tile layer in full.tmx holds.
var wantGIDs = []tmx.GID{1, 2 | 0x80000000, 3 | 0x40000000, 0}

func TestDecodesGym04(t *testing.T) {
	m := decodeFile[tmx.Map](t, assets, "tiled/gym04.tmx")
	require.NoError(t, m.Validate())

	require.Equal(t, tmx.Orthogonal, m.Orientation)
	require.Equal(t, 10, m.Width)
	require.Equal(t, 6, m.Height)
	require.Len(t, m.Tilesets, 2)
	require.Len(t, m.Layers, 7)

	for l := range m.Layers.All() {
		if l.Tiles == nil {
			continue
		}
		gids, err := l.Tiles.Data.GIDs()
		require.NoError(t, err)
		require.Len(t, gids, 60, l.Tiles.Name)
	}

	collision := layerByName(t, m, "collision").Objects
	require.Len(t, collision.Objects, 6)
	require.Equal(t, tmx.ShapeRectangle, collision.Objects[0].Shape())
}

func TestDecodesTheMap(t *testing.T) {
	m := decodeFile[tmx.Map](t, testdata, "full.tmx")
	require.NoError(t, m.Validate())

	require.Equal(t, "level", m.Class)
	require.Equal(t, tmx.Hexagonal, m.Orientation)
	require.Equal(t, tmx.RightDown, m.RenderOrder)
	require.Equal(t, 7, m.HexSideLength)
	require.Equal(t, tmx.StaggerY, m.StaggerAxis)
	require.Equal(t, tmx.StaggerOdd, m.StaggerIndex)
	require.Equal(t, 4.0, m.ParallaxOriginX)
	require.Equal(t, 5.0, m.ParallaxOriginY)
	require.Equal(t, &tmx.Color{A: 0x80, R: 0x11, G: 0x22, B: 0x33}, m.BackgroundColor)

}

func TestDecodesProperties(t *testing.T) {
	props := decodeFile[tmx.Map](t, testdata, "full.tmx").Properties
	get := func(name string) tmx.Property {
		p, ok := props.Get(name)
		require.True(t, ok, name)
		return p
	}

	require.Equal(t, tmx.PropertyString, get("music").Type)
	require.Equal(t, "calm", get("music").Value)
	require.Equal(t, "line one\nline two", get("intro").Value)
	require.Equal(t, "next.tmx", get("next").Value)

	f, err := get("speed").Float()
	require.NoError(t, err)
	require.Equal(t, 1.5, f)

	n, err := get("lives").Int()
	require.NoError(t, err)
	require.Equal(t, 3, n)

	b, err := get("hard").Bool()
	require.NoError(t, err)
	require.True(t, b)

	c, err := get("tint").Color()
	require.NoError(t, err)
	require.Equal(t, tmx.Color{A: 0xff, G: 0xff}, c)

	id, err := get("door").ObjectID()
	require.NoError(t, err)
	require.Equal(t, 7, id)

	spawn := get("spawn")
	require.Equal(t, tmx.PropertyClass, spawn.Type)
	require.Equal(t, "Spawn", spawn.PropertyType)
	count, ok := spawn.Members.Get("count")
	require.True(t, ok)
	require.Equal(t, "2", count.Value)

	_, ok = props.Get("missing")
	require.False(t, ok)
}

func TestKeepsLayerOrder(t *testing.T) {
	m := decodeFile[tmx.Map](t, testdata, "full.tmx")

	var names []string
	for l := range m.Layers.All() {
		names = append(names, l.Base().Name)
	}
	require.Equal(t, []string{"csv", "encoded", "raw", "zlib", "nested", "gzip", "xml", "things", "sky", "empty"}, names)

	require.NotNil(t, m.Layers[1].Group)
	require.Len(t, m.Layers[1].Group.Layers, 3)
	require.Equal(t, 0.75, m.Layers[1].Group.Opacity)
}

func TestDecodesTileLayers(t *testing.T) {
	m := decodeFile[tmx.Map](t, testdata, "full.tmx")

	l := layerByName(t, m, "csv").Tiles
	require.Equal(t, 1, l.ID)
	require.Equal(t, "ground", l.Class)
	require.Equal(t, 0.5, l.Opacity)
	require.False(t, l.Visible)
	require.Equal(t, &tmx.Color{A: 0xff, R: 0xff}, l.TintColor)
	require.Equal(t, 1.5, l.OffsetX)
	require.Equal(t, -2.0, l.OffsetY)
	require.Equal(t, 0.5, l.ParallaxX)
	require.Equal(t, 2.0, l.ParallaxY)
	require.Equal(t, tmx.BlendMultiply, l.Mode)
	solid, _ := l.Properties.Get("solid")
	require.Equal(t, tmx.PropertyBool, solid.Type)

	defaults := layerByName(t, m, "raw").Tiles
	require.Equal(t, 1.0, defaults.Opacity)
	require.True(t, defaults.Visible)
	require.Equal(t, 1.0, defaults.ParallaxX)
	require.Equal(t, tmx.BlendNormal, defaults.Mode)
	require.Nil(t, defaults.TintColor)

	for _, name := range []string{"csv", "raw", "zlib", "gzip", "xml"} {
		gids, err := layerByName(t, m, name).Tiles.Data.GIDs()
		require.NoError(t, err, name)
		require.Equal(t, wantGIDs, gids, name)
	}
}

func TestGIDFlags(t *testing.T) {
	g := tmx.GID(0xF0000000 | 7)
	require.Equal(t, uint32(7), g.ID())
	require.True(t, g.FlippedHorizontally())
	require.True(t, g.FlippedVertically())
	require.True(t, g.FlippedDiagonally())
	require.True(t, g.RotatedHex120())

	plain := tmx.GID(7)
	require.Equal(t, uint32(7), plain.ID())
	require.False(t, plain.FlippedHorizontally())
	require.False(t, plain.FlippedVertically())
	require.False(t, plain.FlippedDiagonally())
	require.False(t, plain.RotatedHex120())
}

func TestDecodesChunks(t *testing.T) {
	m := decodeFile[tmx.Map](t, testdata, "infinite.tmx")
	require.NoError(t, m.Validate())
	require.True(t, m.Infinite)
	require.Equal(t, tmx.LeftUp, m.RenderOrder)

	data := m.Layers[0].Tiles.Data
	require.Len(t, data.Chunks, 2)
	require.Equal(t, -16, data.Chunks[0].X)

	gids, err := data.ChunkGIDs(data.Chunks[1])
	require.NoError(t, err)
	require.Equal(t, []tmx.GID{5, 6, 7, 8}, gids)
}

func TestDecodesObjects(t *testing.T) {
	objects := layerByName(t, decodeFile[tmx.Map](t, testdata, "full.tmx"), "things").Objects
	require.Equal(t, tmx.DrawIndex, objects.DrawOrder)
	require.Equal(t, 0.9, objects.Opacity)

	o := objects.Objects
	require.Equal(t, tmx.ShapeRectangle, o[0].Shape())
	require.Equal(t, "crate", o[0].ClassName())
	require.Equal(t, 1.5, o[0].X)
	require.Equal(t, 2.5, o[0].Y)
	require.Equal(t, 45.0, o[0].Rotation)
	require.Equal(t, 1.0, o[0].Opacity)
	require.True(t, o[0].Visible)

	require.Equal(t, tmx.ShapeEllipse, o[1].Shape())
	require.Equal(t, "enemy", o[1].ClassName())
	require.Equal(t, tmx.ShapeCapsule, o[2].Shape())
	require.Equal(t, tmx.ShapePoint, o[3].Shape())

	require.Equal(t, tmx.ShapePolygon, o[4].Shape())
	require.Equal(t, tmx.Points{{0, 0}, {16, 0}, {8, -12.5}}, o[4].Polygon.Points)
	require.Equal(t, tmx.ShapePolyline, o[5].Shape())
	require.Equal(t, tmx.Points{{0, 0}, {4, 4}}, o[5].Polyline.Points)

	require.Equal(t, tmx.ShapeText, o[6].Shape())
	require.Equal(t, &tmx.Text{
		Content:    "Hello",
		FontFamily: "Mono",
		HAlign:     tmx.HAlignCenter,
		VAlign:     tmx.VAlignBottom,
		Color:      tmx.Color{A: 0xff, R: 0x11, G: 0x22, B: 0x33},
		PixelSize:  12,
		Wrap:       true,
		Bold:       true,
		Italic:     true,
		Underline:  true,
		Strikeout:  true,
	}, o[6].Text)

	require.Equal(t, tmx.ShapeTile, o[7].Shape())
	require.Equal(t, uint32(50), o[7].GID.ID())
	require.True(t, o[7].GID.FlippedHorizontally())
	require.Equal(t, 0.5, o[7].Opacity)
	require.False(t, o[7].Visible)

	require.Equal(t, "crate.tx", o[8].Template)
	hp, _ := o[8].Properties.Get("hp")
	require.Equal(t, "5", hp.Value)
}

func TestTextDefaults(t *testing.T) {
	m, err := decode(t, `<map><objectgroup><object><text>Hi</text></object></objectgroup></map>`)
	require.NoError(t, err)
	require.Equal(t, &tmx.Text{
		Content:    "Hi",
		FontFamily: "sans-serif",
		HAlign:     tmx.HAlignLeft,
		VAlign:     tmx.VAlignTop,
		Color:      tmx.Color{A: 0xff},
		PixelSize:  16,
		Kerning:    true,
	}, m.Layers[0].Objects.Objects[0].Text)
}

func TestDecodesImageLayers(t *testing.T) {
	m := decodeFile[tmx.Map](t, testdata, "full.tmx")

	sky := layerByName(t, m, "sky").Image
	require.True(t, sky.RepeatX)
	require.True(t, sky.RepeatY)
	require.Equal(t, "sky.png", sky.Image.Source)
	require.Equal(t, 320, sky.Image.Width)

	require.Nil(t, layerByName(t, m, "empty").Image.Image)
}

func TestSkipsUnknownElements(t *testing.T) {
	m, err := decode(t, `<map><future/><layer name="a"/><group><future/><imagelayer name="b"/></group></map>`)
	require.NoError(t, err)
	require.Len(t, m.Layers, 2)
	require.Len(t, m.Layers[1].Group.Layers, 1)
}

func TestRejectsBadValues(t *testing.T) {
	for name, src := range map[string]string{
		"other root":  `<tileset/>`,
		"firstgid":    `<map><tileset firstgid="x" source="a.tsx"/></map>`,
		"infinite":    `<map infinite="2"/>`,
		"color":       `<map backgroundcolor="#12"/>`,
		"color hex":   `<map backgroundcolor="#zzzzzz"/>`,
		"points":      `<map><objectgroup><object><polygon points="1 2"/></object></objectgroup></map>`,
		"point value": `<map><objectgroup><object><polygon points="1,x"/></object></objectgroup></map>`,
	} {
		_, err := decode(t, src)
		require.Error(t, err, name)
	}
}

func TestRejectsBadTileData(t *testing.T) {
	for name, data := range map[string]tmx.LayerData{
		"csv":    {Encoding: tmx.EncodingCSV, Data: "1,x"},
		"base64": {Encoding: tmx.EncodingBase64, Data: "!!"},
		"length": {Encoding: tmx.EncodingBase64, Data: "AQID"},
		"zlib":   {Encoding: tmx.EncodingBase64, Compression: tmx.CompressionZlib, Data: "AQAAAA=="},
		"gzip":   {Encoding: tmx.EncodingBase64, Compression: tmx.CompressionGzip, Data: "AQAAAA=="},
	} {
		_, err := data.GIDs()
		require.ErrorIs(t, err, errs.ErrFormat, name)
	}

	_, err := tmx.LayerData{Encoding: tmx.EncodingBase64, Compression: tmx.CompressionZstd, Data: "AQAAAA=="}.GIDs()
	require.ErrorIs(t, err, errs.ErrNotSupported)
}

const validMap = `<map orientation="orthogonal" width="2" height="1" tilewidth="16" tileheight="16" %s>%s</map>`

const validLayer = `<layer id="1" width="2" height="1"><data encoding="csv">1,2</data></layer>`

func TestValidateMap(t *testing.T) {
	m, err := decode(t, `<map orientation="flat" renderorder="up"/>`)
	require.NoError(t, err)

	err = m.Validate()
	require.ErrorIs(t, err, errs.ErrArgument)
	require.ErrorIs(t, err, errs.ErrArgumentOutOfRange)
	for _, field := range []string{`"flat"`, `"up"`, "width", "height", "tilewidth", "tileheight"} {
		require.ErrorContains(t, err, field)
	}

	hex, err := decode(t, fmt.Sprintf(validMap, "", ""))
	require.NoError(t, err)
	hex.Orientation = tmx.Hexagonal
	require.ErrorContains(t, hex.Validate(), "staggeraxis")
	require.ErrorContains(t, hex.Validate(), "staggerindex")
}

func TestValidate(t *testing.T) {
	ok, err := decode(t, fmt.Sprintf(validMap, "", validLayer))
	require.NoError(t, err)
	require.NoError(t, ok.Validate())

	cases := map[string]struct {
		attrs string
		body  string
		want  error
	}{
		"property value":     {"", `<properties><property name="n" type="int" value="x"/></properties>`, errs.ErrArgument},
		"property type":      {"", `<properties><property name="n" type="list"/></properties>`, errs.ErrArgument},
		"class member":       {"", `<properties><property name="c" type="class"><properties><property name="n" type="bool" value="maybe"/></properties></property></properties>`, errs.ErrArgument},
		"firstgid order":     {"", `<tileset firstgid="5" source="a.tsx"/><tileset firstgid="5" source="b.tsx"/>`, errs.ErrArgumentOutOfRange},
		"firstgid missing":   {"", `<tileset source="a.tsx"/>`, errs.ErrArgumentOutOfRange},
		"csv compressed":     {"", `<layer id="1" width="2" height="1"><data encoding="csv" compression="zlib">1,2</data></layer>`, errs.ErrArgument},
		"unknown encoding":   {"", `<layer id="1" width="2" height="1"><data encoding="hex">1,2</data></layer>`, errs.ErrArgument},
		"empty data":         {"", `<layer id="1" width="2" height="1"><data encoding="csv"> </data></layer>`, errs.ErrArgument},
		"chunks when finite": {"", `<layer id="1" width="2" height="1"><data encoding="csv"><chunk width="1" height="1">1</chunk></data></layer>`, errs.ErrArgument},
		"data when infinite": {`infinite="1"`, validLayer, errs.ErrArgument},
		"chunk size":         {`infinite="1"`, `<layer id="1" width="2" height="1"><data encoding="csv"><chunk>1</chunk></data></layer>`, errs.ErrArgumentOutOfRange},
		"layer size":         {"", `<layer id="1"><data encoding="csv">1,2</data></layer>`, errs.ErrArgumentOutOfRange},
		"blend mode":         {"", `<layer id="1" width="2" height="1" mode="glow"><data encoding="csv">1,2</data></layer>`, errs.ErrArgument},
		"opacity":            {"", `<group opacity="2"/>`, errs.ErrArgumentOutOfRange},
		"nested opacity":     {"", `<group><objectgroup opacity="-1"/></group>`, errs.ErrArgumentOutOfRange},
		"draw order":         {"", `<objectgroup draworder="random"/>`, errs.ErrArgument},
		"empty polygon":      {"", `<objectgroup><object><polygon points=""/></object></objectgroup>`, errs.ErrArgument},
		"negative size":      {"", `<objectgroup><object width="-1"/></objectgroup>`, errs.ErrArgumentOutOfRange},
		"text align":         {"", `<objectgroup><object><text halign="middle"/></object></objectgroup>`, errs.ErrArgument},
		"text size":          {"", `<objectgroup><object><text pixelsize="0"/></object></objectgroup>`, errs.ErrArgumentOutOfRange},
		"image source":       {"", `<imagelayer><image/></imagelayer>`, errs.ErrArgument},
		"tile size":          {"", `<tileset firstgid="1"><image source="a.png"/></tileset>`, errs.ErrArgumentOutOfRange},
		"collection image":   {"", `<tileset firstgid="1"><tile id="0"/></tileset>`, errs.ErrArgument},
		"alignment":          {"", `<tileset firstgid="1" objectalignment="middle"/>`, errs.ErrArgument},
		"render size":        {"", `<tileset firstgid="1" tilerendersize="huge"/>`, errs.ErrArgument},
		"fill mode":          {"", `<tileset firstgid="1" fillmode="tile"/>`, errs.ErrArgument},
		"spacing":            {"", `<tileset firstgid="1" spacing="-1"/>`, errs.ErrArgumentOutOfRange},
		"frame":              {"", `<tileset firstgid="1"><image source="a.png"/><tile id="0"><animation><frame duration="-1"/></animation></tile></tileset>`, errs.ErrArgumentOutOfRange},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			m, err := decode(t, fmt.Sprintf(validMap, tc.attrs, tc.body))
			require.NoError(t, err)
			require.ErrorIs(t, m.Validate(), tc.want)
		})
	}
}
