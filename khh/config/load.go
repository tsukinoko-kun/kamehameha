package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

func Load(p string) (*Config, error) {
	if _, err := os.Stat(p); err != nil {
		return nil, errors.Join(fmt.Errorf("failed to stat file %s", p), err)
	}

	f, err := os.Open(p)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("failed to open file %s", p), err)
	}
	defer f.Close()

	d, err := getDecoderFromExtension(p, f)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("failed to get decoder for file %s", p), err)
	}

	c := new(config)
	if err := d.Decode(c); err != nil {
		return nil, err
	}

	return c.Parse()
}

func New(p string) error {
	f, err := os.Create(p)
	if err != nil {
		return errors.Join(fmt.Errorf("failed to create file %s", p), err)
	}

	e, err := getEncoderFromExtension(p, f)
	if err != nil {
		return errors.Join(fmt.Errorf("failed to get encoder for file %s", p), err)
	}

	c := &config{
		Services: []service{
			{
				Name:  "nginx_service",
				Image: "nginx:latest",
			},
		},
		Journeys: []journey{
			{
				Name:  "get index",
				Steps: []string{"GET http://nginx_service/ --status 200"},
			},
			{
				Name: "get unknown",
				Steps: []string{
					"http get http://nginx_service/unknown --status 404",
					"http head http://nginx_service/unknown --status 404",
				},
			},
		},
		Stages: []stage{
			{
				Name:           "default",
				Clients:        1,
				Duration:       "1s",
				DiskCorruption: "5%",
				NetworkFailure: "23.4%",
				FullOutage:     "2.01%",
			},
		},
	}

	if err := e.Encode(c); err != nil {
		return errors.Join(fmt.Errorf("failed to encode config to file %s", p), err)
	}

	return nil
}

func getDecoderFromExtension(p string, f io.Reader) (Decoder, error) {
	switch filepath.Ext(p) {
	case ".json", ".json5", ".jsonc":
		return json.NewDecoder(f), nil
	case ".yaml", ".yml":
		return yaml.NewDecoder(f), nil
	case ".toml":
		return &myTomlDecoder{td: toml.NewDecoder(f)}, nil
	default:
		return nil, fmt.Errorf("unsupported file extension")
	}
}

func getEncoderFromExtension(p string, f io.Writer) (Encoder, error) {
	switch filepath.Ext(p) {
	case ".json", ".json5", ".jsonc":
		e := json.NewEncoder(f)
		e.SetIndent("", "    ")
		return e, nil
	case ".yaml", ".yml":
		return yaml.NewEncoder(f), nil
	case ".toml":
		return toml.NewEncoder(f), nil
	default:
		return nil, fmt.Errorf("unsupported file extension")
	}
}
