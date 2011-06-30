package file

import (
	"os"
	"syscall"
	"fmt"
	)	

type File struct {
	fd int
	name string
}

func newFile(fd int, name string) *File {
	if fd < 0 {
		return nil
	}
	return &File{fd,name}
}

func OpenFile(name string, mode int, perm uint32) (file *File, err os.Error) {
	r, e := syscall.Open(name, mode, perm)
	if e != 0 {
		err = os.Errno(e)
	}
	return newFile(r, name), err
}

const (
	O_RDONLY = syscall.O_RDONLY
	O_RDWR   = syscall.O_RDWR
	O_CREATE = syscall.O_CREAT
	O_TRUNC  = syscall.O_TRUNC
)

func Open(name string) (file *File, err os.Error) {
	return OpenFile(name, O_RDONLY, 0)
}

func Create(name string) (file *File, err os.Error) {
	return OpenFile(name, O_RDWR|O_CREATE|O_TRUNC, 0666)
}

func (file *File) Close() os.Error {
	if file == nil {
		return os.EINVAL
	}
	e := syscall.Close(file.fd)
	file.fd = -1 // so it can't be closed again
	if e != 0 {
		return os.Errno(e)
	}
	return nil
}

func (file *File) Read(b []byte) (ret int, err os.Error) {
	fmt.Printf("Called file.Read(%s)\n", len(b));
	if file == nil {
		return -1, os.EINVAL
	}
	r, e := syscall.Read(file.fd, b)
	if e != 0 {
		err = os.Errno(e)
	}
	return int(r), err
}

func (file *File) Write(b []byte) (ret int, err os.Error) {
	if file == nil {
		return -1, os.EINVAL
        }
        r, e := syscall.Write(file.fd, b)
        if e != 0 {
		err = os.Errno(e)
	}
        return int(r), err
}

func (file *File) Seek(offset int64, whence int) (ret int, err os.Error) {
	if file == nil {
		return -1, os.EINVAL
	}
	r, e := syscall.Seek(file.fd, offset, whence)
	if e != 0 {
		err = os.Errno(e)
	}
	return int(r), err
}

func (file *File) String() string {
        return file.name
}