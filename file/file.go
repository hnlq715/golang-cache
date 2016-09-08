package file

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type File struct {
	option *Option
}

type Option struct {
	DirName string
}

func New(option *Option) *File {
	return &File{
		option: option,
	}
}

func (f *File) Get(key string) ([]byte, error) {
	filename := filepath.Join(f.option.DirName, generateFileName(key))

	_, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return data, err
}

func (f *File) Set(key string, data io.Reader) error {
	filename, _ := filepath.Abs(filepath.Join(f.option.DirName, generateFileName(key)))

	dir := filepath.Dir(filename)
	err := os.MkdirAll(dir, 0600)
	if err != nil {
		return err
	}

	tmp, err := ioutil.TempFile(dir, filepath.Base(filename)+".tmp")
	if err != nil {
		return err
	}

	if _, err = io.Copy(tmp, data); err != nil {
		return err
	}

	if err = tmp.Close(); err != nil {
		return err
	}

	if err = os.Rename(tmp.Name(), filename); err != nil {
		return err
	}

	return nil
}

func (f *File) Delete(key string) error {
	filename := filepath.Join(f.option.DirName, generateFileName(key))

	if err := os.Remove(filename); err != nil {
		return err
	}

	return nil
}

func (f *File) DirName() string {
	return f.option.DirName
}

func generateFileName(key string) string {
	name := hex.EncodeToString([]byte(fmt.Sprintf("%s", md5.Sum([]byte(key)))))

	level1 := name[31:32]
	level2 := name[29:31]

	return filepath.Join(level1, level2, name)
}
