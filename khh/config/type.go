package config

import (
	"errors"
	"time"

	"github.com/BurntSushi/toml"
)

type (
	config struct {
		Services []service `json:"services" yaml:"services" toml:"services"`
		Journeys []journey `json:"journeys" yaml:"journeys" toml:"journeys"`
		Stages   []stage   `json:"stages" yaml:"stages" toml:"stages"`
	}

	service struct {
		Name  string `json:"name" yaml:"name" toml:"name"`
		Image string `json:"image" yaml:"image" toml:"image"`
	}

	journey struct {
		Name  string   `json:"name" yaml:"name" toml:"name"`
		Steps []string `json:"steps" yaml:"steps" toml:"steps"`
	}

	stage struct {
		Name           string `json:"name" yaml:"name" toml:"name"`
		Clients        int    `json:"clients" yaml:"clients" toml:"clients"`
		Duration       string `json:"duration" yaml:"duration" toml:"duration"`
		DiskCorruption string `json:"disk_corruption" yaml:"disk_corruption" toml:"disk_corruption"`
		NetworkFailure string `json:"network_failure" yaml:"network_failure" toml:"network_failure"`
		FullOutage     string `json:"full_outage" yaml:"full_outage" toml:"full_outage"`
	}

	Config struct {
		Services []Service
		Journeys []Journey
		Stages   []Stage
	}

	Service struct {
		Name  string
		Image string
	}

	Journey struct {
		Name  string
		Steps []string
	}

	Stage struct {
		Name           string
		Clients        int
		Duration       time.Duration
		DiskCorruption Probability
		NetworkFailure Probability
		FullOutage     Probability
	}
)

type (
	Decoder interface {
		Decode(v interface{}) error
	}

	Encoder interface {
		Encode(v interface{}) error
	}

	myTomlDecoder struct {
		td *toml.Decoder
	}
)

func (c *config) Parse() (*Config, error) {
	cfg := &Config{
		Stages: make([]Stage, 0, len(c.Stages)),
	}

	for _, s := range c.Services {
		service, err := s.Parse()
		if err != nil {
			return nil, errors.Join(errors.New("failed to parse service"), err)
		}
		cfg.Services = append(cfg.Services, *service)
	}

	for _, j := range c.Journeys {
		journey, err := j.Parse()
		if err != nil {
			return nil, errors.Join(errors.New("failed to parse journey"), err)
		}
		cfg.Journeys = append(cfg.Journeys, *journey)
	}

	for _, s := range c.Stages {
		stage, err := s.Parse()
		if err != nil {
			return nil, errors.Join(errors.New("failed to parse stage"), err)
		}
		cfg.Stages = append(cfg.Stages, *stage)
	}

	return cfg, nil
}

func (s *service) Parse() (*Service, error) {
	return &Service{
		Name:  s.Name,
		Image: s.Image,
	}, nil
}

func (j *journey) Parse() (*Journey, error) {
	return &Journey{
		Name:  j.Name,
		Steps: j.Steps,
	}, nil
}

func (s *stage) Parse() (*Stage, error) {
	duration, err := time.ParseDuration(s.Duration)
	if err != nil {
		return nil, errors.Join(errors.New("failed to parse duration"), err)
	}

	diskCorruption, err := parseProbability(s.DiskCorruption)
	if err != nil {
		return nil, errors.Join(errors.New("failed to parse disk corruption"), err)
	}

	networkFailure, err := parseProbability(s.NetworkFailure)
	if err != nil {
		return nil, errors.Join(errors.New("failed to parse network failure"), err)
	}

	fullOutage, err := parseProbability(s.FullOutage)
	if err != nil {
		return nil, errors.Join(errors.New("failed to parse full outage"), err)
	}

	return &Stage{
		Name:           s.Name,
		Clients:        s.Clients,
		Duration:       duration,
		DiskCorruption: diskCorruption,
		NetworkFailure: networkFailure,
		FullOutage:     fullOutage,
	}, nil
}

func (d *myTomlDecoder) Decode(v interface{}) error {
	_, err := d.td.Decode(v)
	return err
}
