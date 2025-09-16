package util

import (
	"encoding/gob"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
)

// StoreJson 缓存结构体到指定json文件中
func StoreJson(v interface{}, cacheFilePath string) error {
	cachePath, err := FilePath(filepath.Dir(cacheFilePath), filepath.Base(cacheFilePath))
	if err != nil {
		return err
	}

	file, err := os.Create(cachePath)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := jsoniter.NewEncoder(file).Encode(v); err != nil {
		return err
	}

	return nil
}

// LoadJson 加载json文件内容到结构体中
func LoadJson(v interface{}, cacheFilePath string) error {
	if !FileExists(cacheFilePath) {
		return errors.Errorf("file [%s] doesn't exists", cacheFilePath)
	}

	file, err := os.Open(cacheFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := jsoniter.NewDecoder(file).Decode(v); err != nil {
		return err
	}
	return nil
}

// StoreGob 缓存结构体到指定二进制文件中
func StoreGob(v interface{}, cacheFilePath string) error {
	cachePath, err := FilePath(filepath.Dir(cacheFilePath), filepath.Base(cacheFilePath))
	if err != nil {
		return err
	}

	file, err := os.Create(cachePath)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := gob.NewEncoder(file).Encode(v); err != nil {
		return err
	}

	return nil
}

// LoadGob 加载二进制文件内容到结构体中
func LoadGob(v interface{}, cacheFilePath string) error {
	if !FileExists(cacheFilePath) {
		return errors.Errorf("file [%s] doesn't exists", cacheFilePath)
	}

	file, err := os.Open(cacheFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := gob.NewDecoder(file).Decode(v); err != nil {
		return err
	}
	return nil
}
