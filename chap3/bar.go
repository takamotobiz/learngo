package osmpbfparser

import (
	"github.com/vbauerster/mpb/v5"
	"github.com/vbauerster/mpb/v5/decor"
	"sync"
)

// newProgress return new mpb progress with input total.
func newProgress(n int) (*mpb.Progress, *sync.WaitGroup) {
	var wg sync.WaitGroup
	wg.Add(n)
	return mpb.New(mpb.WithWaitGroup(&wg), mpb.WithWidth(100)), &wg
}

func addBar(pb *mpb.Progress, name string, total int) *mpb.Bar {
	return pb.AddBar(
		int64(total),
		mpb.BarStyle("╢▌▌░╟"),
		mpb.PrependDecorators(
			decor.Name(name, decor.WC{W: 20, C: decor.DidentRight}),
		),
		mpb.AppendDecorators(
			decor.CountersNoUnit("%d/%d ", decor.WCSyncWidth),
			decor.OnComplete(
				decor.AverageETA(decor.ET_STYLE_GO, decor.WCSyncSpaceR),
				"done",
			),
			decor.Percentage(),
		),
	)
}
