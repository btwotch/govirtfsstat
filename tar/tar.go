package tar

import (
	"archive/tar"
	"errors"
	"fmt"
	"govirtfsstat/stat"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

func Untar(tarReader io.Reader, destPath string) {
	tr := tar.NewReader(tarReader)

	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}

		fmt.Printf("-- %+v\n", hdr)

		path := filepath.Join(destPath, hdr.Name)
		if !strings.HasPrefix(path, destPath) {
			path = filepath.Join(destPath, filepath.Base(hdr.Name))
		}

		mode := hdr.FileInfo().Mode()

		if hdr.Typeflag == tar.TypeDir || hdr.FileInfo().IsDir() {
			err := os.MkdirAll(path, 0700)
			if err != nil {
				panic(err)
			}
			mode = mode | syscall.S_IFDIR
			// } else if hdr.Typeflag == tar.TypeSymlink {
			// 	mode = mode | syscall.S_IFLNK
			// 	err := os.MkdirAll(filepath.Dir(path), 0700)
			// 	if err != nil {
			// 		panic(err)
			// 	}
			// 	err = os.Symlink(hdr.Linkname, path)
			// 	if err != nil {
			// 		panic(err)
			// 	}
		} else {
			err := os.MkdirAll(filepath.Dir(path), 0700)
			if err != nil {
				panic(err)
			}
			fp, err := os.OpenFile(path, syscall.O_CREAT|syscall.O_WRONLY, 0600)
			if err != nil {
				panic(err)
			}
			_, err = io.Copy(fp, tr)
			if err != nil {
				panic(err)
			}
			fp.Close()
		}

		if hdr.Typeflag == tar.TypeBlock || hdr.Typeflag == tar.TypeChar {
			dev := unix.Mkdev(uint32(hdr.Devmajor), uint32(hdr.Devminor))
			err = stat.SetRDev(path, dev)
			if err != nil {
				panic(err)
			}
		}

		if hdr.Typeflag == tar.TypeChar {
			mode = mode | syscall.S_IFCHR
		}
		if hdr.Typeflag == tar.TypeBlock {
			mode = mode | syscall.S_IFBLK
		}
		if hdr.Typeflag == tar.TypeLink {
			mode = mode | syscall.S_IFREG
		}
		if hdr.Typeflag == tar.TypeFifo {
			mode = mode | syscall.S_IFIFO
		}
		if hdr.Typeflag == tar.TypeSymlink {
			mode = mode | syscall.S_IFLNK
			err := os.WriteFile(path, []byte(hdr.Linkname), 0600)
			if err != nil {
				panic(err)
			}
		}

		err = stat.SetMode(path, uint32(mode))
		if err != nil {
			panic(err)
		}
		err = stat.SetUid(path, uint32(hdr.Uid))
		if err != nil {
			panic(err)
		}
		err = stat.SetGid(path, uint32(hdr.Gid))
		if err != nil {
			panic(err)
		}
		fmt.Printf("tar sys: %+v\n", hdr.FileInfo().Sys())
	}
}
