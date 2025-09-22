package stat

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"syscall"
)

func Stat(path string) (syscall.Stat_t, error) {
	sz, err := syscall.Listxattr(path, []byte{})
	if err != nil {
		return syscall.Stat_t{}, fmt.Errorf("listing extended attributes failed: %w", err)
	}
	bs := make([]byte, sz)
	_, err = syscall.Listxattr(path, bs)

	virtfsKeys := map[string]bool{
		"user.virtfs.uid":  false,
		"user.virtfs.gid":  false,
		"user.virtfs.mode": false,
	}
	bss := bytes.Split(bs, []byte{0})
	for _, bs := range bss {
		_, found := virtfsKeys[string(bs)]
		if found {
			virtfsKeys[string(bs)] = true
		}
	}

	for _, val := range virtfsKeys {
		if !val {
			// not a virtfs file
			var st syscall.Stat_t
			err := syscall.Stat(path, &st)
			if err != nil {
				return syscall.Stat_t{}, fmt.Errorf("syscall.Stat failed: %w", err)
			}

			return st, nil
		}
	}

	return VirtfsStat(path)
}
func extractId(path string, key string) (uint32, error) {
	bs := make([]byte, 4)
	_, err := syscall.Getxattr(path, key, bs)
	if err != nil {
		return 0, fmt.Errorf("getting attribute '%s' from '%s' failed: %w", key, path, err)
	}

	u := binary.LittleEndian.Uint32(bs)
	return u, nil
}

func VirtfsStat(path string) (syscall.Stat_t, error) {
	mode, err := extractMode(path)
	if err != nil {
		return syscall.Stat_t{}, fmt.Errorf("virtfsstat failed to extract mode: %w", err)
	}
	uid, err := extractId(path, "user.virtfs.uid")
	if err != nil {
		return syscall.Stat_t{}, fmt.Errorf("virtfsstat failed to extract uid: %w", err)
	}
	gid, err := extractId(path, "user.virtfs.gid")
	if err != nil {
		return syscall.Stat_t{}, fmt.Errorf("virtfsstat failed to extract gid: %w", err)
	}

	var st syscall.Stat_t
	syscall.Stat(path, &st)
	st.Mode = mode
	st.Uid = uid
	st.Gid = gid
	st.Rdev, err = extractRdev(path)
	if err != nil {
		return syscall.Stat_t{}, fmt.Errorf("virtfsstat failed to extract rdev: %w", err)
	}

	return st, nil
}

func extractRdev(path string) (uint64, error) {
	key := "user.virtfs.rdev"

	sz, err := syscall.Getxattr(path, key, []byte{})
	if err != nil {
		return 0, nil
	}
	bs := make([]byte, sz)

	_, err = syscall.Getxattr(path, key, bs)
	if err != nil {
		return 0, fmt.Errorf("extracting actual rdev value failed: %w", err)
	}

	u := binary.LittleEndian.Uint64(bs[0:8])

	return u, nil
}

func extractMode(path string) (uint32, error) {
	key := "user.virtfs.mode"

	sz, err := syscall.Getxattr(path, key, []byte{})
	if err != nil {
		return 0, fmt.Errorf("extracting '%s' failed: %w", key, err)
	}
	bs := make([]byte, sz)

	_, err = syscall.Getxattr(path, key, bs)
	if err != nil {
		return 0, fmt.Errorf("extracting '%s' failed: %w", key, err)
	}

	u := binary.LittleEndian.Uint32(bs[0:4])
	return u, nil
}

func SetMode(path string, mode uint32) error {
	key := "user.virtfs.mode"

	data := binary.LittleEndian.AppendUint32([]byte{}, mode)
	return syscall.Setxattr(path, key, data, 0)
}

func SetRDev(path string, rdev uint64) error {
	key := "user.virtfs.rdev"

	data := binary.LittleEndian.AppendUint64([]byte{}, rdev)
	return syscall.Setxattr(path, key, data, 0)
}

func setId(path string, key string, id uint32) error {
	data := binary.LittleEndian.AppendUint32([]byte{}, id)
	return syscall.Setxattr(path, key, data, 0)
}

func SetUid(path string, uid uint32) error {
	key := "user.virtfs.uid"
	return setId(path, key, uid)
}
func SetGid(path string, uid uint32) error {
	key := "user.virtfs.gid"
	return setId(path, key, uid)
}
