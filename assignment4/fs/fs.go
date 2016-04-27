package fs

import (
	_ "fmt"
	"sync"
	"time"
)

type FileInfo struct {
	filename   string
	contents   []byte
	version    int
	absexptime time.Time
	timer      *time.Timer
}

type FS struct {
	sync.RWMutex
	Dir map[string]*FileInfo
}

var gversion = 0 // global version

func (fi *FileInfo) cancelTimer() {
	if fi.timer != nil {
		fi.timer.Stop()
		fi.timer = nil
	}
}

func (fileStore *FS) ProcessMsg(msg *Msg) *Msg {
	switch msg.Kind {
	case 'r':
		return fileStore.processRead(msg)
	case 'w':
		return fileStore.processWrite(msg)
	case 'c':
		return fileStore.processCas(msg)
	case 'd':
		return fileStore.processDelete(msg)
	}

	// Default: Internal error. Shouldn't come here since
	// the msg should have been validated earlier.
	return &Msg{Kind: 'I'}
}

func (fileStore *FS) processRead(msg *Msg) *Msg {
	fileStore.RLock()
	defer fileStore.RUnlock()
	if fi := fileStore.Dir[msg.Filename]; fi != nil {
		remainingTime := 0
		if fi.timer != nil {
			remainingTime := int(fi.absexptime.Sub(time.Now()))
			if remainingTime < 0 {
				remainingTime = 0
			}
		}
		return &Msg{
			Kind:     'C',
			Filename: fi.filename,
			Contents: fi.contents,
			Numbytes: len(fi.contents),
			Exptime:  remainingTime,
			Version:  fi.version,
		}
	} else {
		return &Msg{Kind: 'F'} // file not found
	}
}

func (fileStore *FS) internalWrite(msg *Msg) *Msg {
	fi := fileStore.Dir[msg.Filename]
	if fi != nil {
		fi.cancelTimer()
	} else {
		fi = &FileInfo{}
	}

	gversion += 1
	fi.filename = msg.Filename
	fi.contents = msg.Contents
	fi.version = gversion

	var absexptime time.Time
	if msg.Exptime > 0 {
		dur := time.Duration(msg.Exptime) * time.Second
		absexptime = time.Now().Add(dur)
		timerFunc := func(name string, ver int) func() {
			return func() {
				fileStore.processDelete(&Msg{Kind: 'D',
					Filename: name,
					Version:  ver})
			}
		}(msg.Filename, gversion)

		fi.timer = time.AfterFunc(dur, timerFunc)
	}
	fi.absexptime = absexptime
	fileStore.Dir[msg.Filename] = fi

	return fileStore.ok(gversion)
}

func (fileStore *FS) processWrite(msg *Msg) *Msg {
	fileStore.Lock()
	defer fileStore.Unlock()
	return fileStore.internalWrite(msg)
}

func (fileStore *FS) processCas(msg *Msg) *Msg {
	fileStore.Lock()
	defer fileStore.Unlock()

	if fi := fileStore.Dir[msg.Filename]; fi != nil {
		if msg.Version != fi.version {
			return &Msg{Kind: 'V', Version: fi.version}
		}
	}
	return fileStore.internalWrite(msg)
}

func (fileStore *FS) processDelete(msg *Msg) *Msg {
	fileStore.Lock()
	defer fileStore.Unlock()
	fi := fileStore.Dir[msg.Filename]
	if fi != nil {
		if msg.Version > 0 && fi.version != msg.Version {
			// non-zero msg.Version indicates a delete due to an expired timer
			return nil // nothing to do
		}
		fi.cancelTimer()
		delete(fileStore.Dir, msg.Filename)
		return fileStore.ok(0)
	} else {
		return &Msg{Kind: 'F'} // file not found
	}

}

func (fileStore *FS) ok(version int) *Msg {
	return &Msg{Kind: 'O', Version: version}
}
