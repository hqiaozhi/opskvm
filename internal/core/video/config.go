package video

import "github.com/korandiz/v4l"

type Config struct {
	Path   string
	Width  int
	Height int
	FPS    int
	cam    *v4l.Device
}
