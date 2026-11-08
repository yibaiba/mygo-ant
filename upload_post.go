package kit

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/egoist/mygo/ui"
)

// UploadRequest is one file kit (or CustomRequest) sends, as Ant's
// customRequest options. OnProgress / OnSuccess / OnError may run off
// the UI goroutine; kit writes them back on the next frame.
type UploadRequest struct {
	File       UploadFile
	Action     string
	Method     string
	Filename   string
	Data       map[string]string
	Headers    map[string]string
	OnProgress func(percent int)
	OnSuccess  func(resp []byte)
	OnError    func(err error)
}

type uploadPatch struct {
	key     string
	status  UploadStatus
	percent int
	err     string
	hasErr  bool
}

type uploadJobs struct {
	mu      sync.Mutex
	started map[string]bool
	patch   []uploadPatch
}

func (j *uploadJobs) take() []uploadPatch {
	j.mu.Lock()
	defer j.mu.Unlock()
	out := j.patch
	j.patch = nil
	return out
}

func (j *uploadJobs) put(p uploadPatch) {
	j.mu.Lock()
	j.patch = append(j.patch, p)
	j.mu.Unlock()
}

func (j *uploadJobs) forgetMissing(files []UploadFile) {
	have := map[string]bool{}
	for _, f := range files {
		have[f.key()] = true
	}
	j.mu.Lock()
	for k := range j.started {
		if !have[k] {
			delete(j.started, k)
		}
	}
	j.mu.Unlock()
}

func kitUploads(p UploadListProps) bool {
	return p.CustomRequest != nil || p.Action != ""
}

func applyUploadPatches(files []UploadFile, patches []uploadPatch) {
	if len(patches) == 0 {
		return
	}
	idx := map[string]int{}
	for i := range files {
		idx[files[i].key()] = i
	}
	for _, p := range patches {
		i, ok := idx[p.key]
		if !ok {
			continue
		}
		files[i].Status = p.status
		files[i].Percent = p.percent
		if p.hasErr {
			files[i].Error = p.err
		}
	}
}

func startUploads(c *ui.Context, files, added []UploadFile, p UploadListProps, jobs *uploadJobs) {
	if len(added) == 0 || !kitUploads(p) {
		return
	}
	byKey := map[string]UploadFile{}
	for _, f := range files {
		byKey[f.key()] = f
	}
	for _, add := range added {
		f := byKey[add.key()]
		if f.Path == "" {
			f = add
		}
		key := f.key()
		jobs.mu.Lock()
		if jobs.started[key] {
			jobs.mu.Unlock()
			continue
		}
		jobs.started[key] = true
		jobs.mu.Unlock()
		req := UploadRequest{
			File:     f,
			Action:   p.Action,
			Method:   p.Method,
			Filename: p.Filename,
			Data:     p.Data,
			Headers:  p.Headers,
		}
		req.OnProgress = func(percent int) {
			jobs.put(uploadPatch{key: key, status: UploadUploading, percent: percent})
			c.Invalidate()
		}
		req.OnSuccess = func([]byte) {
			jobs.put(uploadPatch{key: key, status: UploadDone, percent: 100})
			c.Invalidate()
		}
		req.OnError = func(err error) {
			msg := "上传失败"
			if err != nil && err.Error() != "" {
				msg = err.Error()
			}
			jobs.put(uploadPatch{key: key, status: UploadError, hasErr: true, err: msg})
			c.Invalidate()
		}
		go runUpload(req, p.CustomRequest)
	}
}

func runUpload(req UploadRequest, custom func(UploadRequest)) {
	if custom != nil {
		custom(req)
		return
	}
	if err := postUpload(req); err != nil {
		if req.OnError != nil {
			req.OnError(err)
		}
		return
	}
}

func postUpload(req UploadRequest) error {
	if req.Action == "" {
		return fmt.Errorf("empty action")
	}
	f, err := os.Open(req.File.Path)
	if err != nil {
		return err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return err
	}
	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)
	go func() {
		err := writeUploadBody(mw, f, st.Size(), req)
		mw.Close()
		if err != nil {
			pw.CloseWithError(err)
			return
		}
		pw.Close()
	}()
	method := req.Method
	if method == "" {
		method = http.MethodPost
	}
	httpReq, err := http.NewRequest(method, req.Action, pr)
	if err != nil {
		pr.Close()
		return err
	}
	httpReq.Header.Set("Content-Type", mw.FormDataContentType())
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		if len(body) > 0 {
			return fmt.Errorf("%s: %s", resp.Status, body)
		}
		return fmt.Errorf("%s", resp.Status)
	}
	if req.OnSuccess != nil {
		req.OnSuccess(body)
	}
	return nil
}

func writeUploadBody(mw *multipart.Writer, r io.Reader, total int64, req UploadRequest) error {
	name := pick(req.Filename == "", "file", req.Filename)
	part, err := mw.CreateFormFile(name, filepath.Base(req.File.Path))
	if err != nil {
		return err
	}
	src := r
	if req.OnProgress != nil {
		req.OnProgress(0)
		src = &progressReader{r: r, total: total, on: req.OnProgress}
	}
	if _, err := io.Copy(part, src); err != nil {
		return err
	}
	for k, v := range req.Data {
		if err := mw.WriteField(k, v); err != nil {
			return err
		}
	}
	if req.OnProgress != nil {
		req.OnProgress(100)
	}
	return nil
}

type progressReader struct {
	r        io.Reader
	n, total int64
	last     int
	on       func(int)
}

func (p *progressReader) Read(b []byte) (int, error) {
	n, err := p.r.Read(b)
	p.n += int64(n)
	if p.on != nil && p.total > 0 {
		pct := int(p.n * 100 / p.total)
		if pct > 100 {
			pct = 100
		}
		if pct != p.last {
			p.last = pct
			p.on(pct)
		}
	}
	return n, err
}
