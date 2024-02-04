package storage

import (
	"encoding/json"
	"github.com/shadyziedan/metrica/internal/models"
	"io"
	"os"
)

type FileStorage struct {
	producer *producer
	consumer *consumer
	fileName string
}

func NewFileStorage(fileName string) *FileStorage {
	return &FileStorage{fileName: fileName}
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

func newProducer(fileName string) (*producer, error) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	return &producer{file: file, encoder: json.NewEncoder(file)}, nil
}

func (fs *FileStorage) ReadMetrics() ([]*models.Metrics, error) {
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
	return fs.producer.encoder.Encode(metric)
}

func (c *consumer) Close() error {
	return c.file.Close()
}

func (p *producer) Close() error {
	return p.file.Close()
}

func (fs *FileStorage) Close() error {
	if fs.producer != nil {
		err := fs.producer.Close()
		if err != nil {
			return err
		}
		fs.producer = nil
	}
	if fs.consumer != nil {
		err := fs.consumer.Close()
		if err != nil {
			return err
		}
		fs.producer = nil
	}
	return nil
}

func (fs *FileStorage) OpenProducer() error {
	p, err := newProducer(fs.fileName)
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
