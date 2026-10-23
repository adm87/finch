package aseprite_test

import (
	"encoding/json"
	"testing"

	"github.com/go-openapi/testify/require"

	"github.com/adm87/finch/pkg/aseprite"
	"github.com/adm87/finch/pkg/errs"
)

func TestFramesUnmarshalArray(t *testing.T) {
	raw := `[
		{"filename": "a", "frame": {"x": 0, "y": 0, "w": 16, "h": 16}, "duration": 100},
		{"filename": "b", "frame": {"x": 16, "y": 0, "w": 16, "h": 16}, "duration": 200}
	]`

	var frames aseprite.Frames
	require.NoError(t, json.Unmarshal([]byte(raw), &frames))
	require.Len(t, frames, 2)
	require.Equal(t, "a", frames[0].FileName)
	require.Equal(t, "b", frames[1].FileName)
	require.Equal(t, 16.0, frames[1].FrameRect.X)
	require.Equal(t, 200, frames[1].Duration)
}

func TestFramesUnmarshalHashKeepsOrderAndNames(t *testing.T) {
	raw := `{
		"run 1.aseprite": {"frame": {"x": 0, "y": 0, "w": 8, "h": 8}, "duration": 50},
		"run 0.aseprite": {"frame": {"x": 8, "y": 0, "w": 8, "h": 8}, "duration": 60}
	}`

	var frames aseprite.Frames
	require.NoError(t, json.Unmarshal([]byte(raw), &frames))
	require.Len(t, frames, 2)
	require.Equal(t, "run 1.aseprite", frames[0].FileName)
	require.Equal(t, "run 0.aseprite", frames[1].FileName)
	require.Equal(t, 60, frames[1].Duration)
}

func TestFramesUnmarshalRejectsOtherLayouts(t *testing.T) {
	for _, raw := range []string{`"frames"`, `42`, `{"a": 1}`} {
		var frames aseprite.Frames
		err := json.Unmarshal([]byte(raw), &frames)
		require.Error(t, err, raw)
		require.ErrorIs(t, err, errs.ErrFormat, raw)
	}
}

func TestAtlasDropMissingTagsAndValidate(t *testing.T) {
	atlas := aseprite.Atlas{
		Frames: make(aseprite.Frames, 3),
		Meta: aseprite.Meta{
			FrameTags: aseprite.FrameTags{
				{Name: "Idle", Direction: aseprite.DirectionForward, From: 0, To: 2},
				{Name: "Gone", Direction: aseprite.DirectionForward, From: 5, To: 6},
			},
		},
	}

	require.Equal(t, []string{"Gone"}, atlas.DropMissingTags())
	require.NotNil(t, atlas.Meta.FrameTags.ByName("Idle"))
	require.Nil(t, atlas.Meta.FrameTags.ByName("Gone"))
	require.NoError(t, atlas.Validate())

	atlas.Meta.FrameTags[0].Direction = "sideways"
	atlas.Meta.FrameTags[0].To = 3
	require.Error(t, atlas.Validate())
}
