package start

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type LengthWriter interface {
	SetContentLength(int64)
	io.Writer
}

type progressWriter struct {
	written int64
	total   int64
	inner   io.Writer
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n, err := pw.inner.Write(p)
	if err != nil {
		return n, err
	}

	pw.written += int64(n)

	percents := int(float64(pw.written) / float64(pw.total) * 100)
	fmt.Printf("\r%d%% [%s%s]", percents, strings.Repeat("#", percents), strings.Repeat(" ", 100-percents))

	if pw.written == pw.total {
		fmt.Printf("\n\n")
		return n, nil
	}

	return n, nil
}

func (pw *progressWriter) SetContentLength(length int64) {
	pw.total = length
}

func download[T LengthWriter](ctx context.Context, url string, dest T) error {
	cli := http.DefaultClient
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := cli.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	dest.SetContentLength(resp.ContentLength)

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	_, err = io.Copy(dest, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
