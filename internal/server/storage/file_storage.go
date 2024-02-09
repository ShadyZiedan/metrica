package storage

import (
	"encoding/json"
	"io"
	"os"
	"sync"

	"github.com/shadyziedan/metrica/internal/models"
)

type FileStorage struct {
	producer *producer
	consumer *consumer
	fileName string
	mode     Mode

	mutex sync.RWMutex
}

func NewFileStorage(fileName string, mode Mode) *FileStorage {
	return &FileStorage{fileName: fileName, mode: mode}
}

type producer struct {
	file    *os.File
	encoder *json.Encoder
}

type consumer struct {
	file    *os.File
	decoder *json.Decoder
}

func newConsumer(fileName string) (*consumer, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	return &consumer{file: file, decoder: json.NewDecoder(file)}, nil
}

type Mode string

const (
	Async  Mode = "async"
	Normal      = "normal"
)

func newProducer(fileName string, mode Mode) (*producer, error) {
	var flag int
	switch mode {
	case Async:
		flag = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	case Normal:
		flag = os.O_WRONLY | os.O_CREATE
	}
	file, err := os.OpenFile(fileName, flag, 0666)
	if err != nil {
		return nil, err
	}
	return &producer{file: file, encoder: json.NewEncoder(file)}, nil
}

func (fs *FileStorage) ReadMetrics() ([]*models.Metrics, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()
	err := fs.OpenConsumer()
	if err != nil {
		return nil, err
	}
	defer func(fs *FileStorage) {
		err := fs.CloseConsumer()
		if err != nil {
			panic(err)
		}
	}(fs)
	var res []*models.Metrics
	for {
		model := &models.Metrics{}
		err := fs.consumer.decoder.Decode(&model)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		res = append(res, model)
	}
	return res, nil
}

func (fs *FileStorage) SaveMetric(metric *models.Metrics) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	err := fs.OpenProducer()
	if err != nil {
		return err
	}
	defer func(producer *producer) {
		err := producer.Close()
		if err != nil {
			panic(err)
		}
	}(fs.producer)
	return fs.producer.encoder.Encode(metric)
}

func (fs *FileStorage) SaveMetrics(metrics []*models.Metrics) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	err := fs.OpenProducer()
	if err != nil {
		return err
	}
	defer func(producer *producer) {
		err := producer.Close()
		if err != nil {
			panic(err)
		}
	}(fs.producer)
	for _, metric := range metrics {
		err := fs.producer.encoder.Encode(metric)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *consumer) Close() error {
	return c.file.Close()
}

func (p *producer) Close() error {
	return p.file.Close()
}

func (fs *FileStorage) CloseProducer() error {
	if fs.producer != nil {
		err := fs.producer.Close()
		if err != nil {
			return err
		}
		fs.producer = nil
	}
	return nil
}
func (fs *FileStorage) CloseConsumer() error {
	if fs.consumer != nil {
		err := fs.consumer.Close()
		if err != nil {
			return err
		}
		fs.producer = nil
	}
	return nil
}

func (fs *FileStorage) Close() error {
	err := fs.CloseProducer()
	if err != nil {
		return err
	}
	return fs.CloseConsumer()
}

func (fs *FileStorage) OpenProducer() error {
	p, err := newProducer(fs.fileName, fs.mode)
	if err != nil {
		return err
	}
	fs.producer = p
	return nil
}

func (fs *FileStorage) OpenConsumer() error {
	c, err := newConsumer(fs.fileName)
	if err != nil {
		return err
	}
	fs.consumer = c
	return nil
}
