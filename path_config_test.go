package main

import (
	"errors"
	"testing"
	"time"

	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
	"github.com/stretchr/testify/assert"
)

func TestConfigReturnError(t *testing.T) {
	var b backend

	config, err := b.Config(GetErrorStorage{})
	assert.NotNil(t, err)
	assert.Nil(t, config)
}

type GetErrorStorage struct {
	logical.Storage
}

func (s GetErrorStorage) Get(path string) (*logical.StorageEntry, error) {
	return nil, errors.New("Fake error")
}

func TestConfigReturnInvalidJSON(t *testing.T) {
	var b backend

	config, err := b.Config(GetInvalidJSONStorage{})
	assert.NotNil(t, err)
	assert.Nil(t, config)
}

type GetInvalidJSONStorage struct {
	logical.Storage
}

func (s GetInvalidJSONStorage) Get(path string) (*logical.StorageEntry, error) {
	return &logical.StorageEntry{
		Key: "config",
		Value: []byte("bad{json,is what I need"),
	}, nil
}

func TestConfigReturnValid(t *testing.T) {
	var b backend
	config, err := b.Config(GetValidStorage{
		Entry: &logical.StorageEntry{
			Key: "config",
			Value: []byte("{\"circleci_token\": \"test-token\", \"ttl\": 300, \"max_ttl\": 900}"),
		},
	})
	assert.Nil(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "test-token", config.CircleCIToken)
	assert.Equal(t, time.Duration(300), config.TTL)
	assert.Equal(t, time.Duration(900), config.MaxTTL)
}

type GetValidStorage struct {
	logical.Storage

	Entry *logical.StorageEntry
}

func (s GetValidStorage) Get(path string) (*logical.StorageEntry, error) {
	return s.Entry, nil
}

func TestParseDurationField(t *testing.T) {
	oneSecond := "1s"
	oneDay := "1d"
	garbage := "garbage"
	blank := ""

	testcases := []struct{
		valueReturned *string
		expectError bool
		expectedValue time.Duration
	} {
		{
			valueReturned: &oneSecond,
			expectError: false,
			expectedValue: time.Second,
		},
		{
			valueReturned: &oneDay,
			expectError: true,
		},
		{
			valueReturned: &garbage,
			expectError: true,
		},
		{
			valueReturned: &blank,
			expectError: false,
			expectedValue: 0,
		},
	}

	for _, tc := range testcases {
		d := &framework.FieldData{
			Raw: map[string]interface{} {
				"key": *tc.valueReturned,
			},
			Schema: map[string]*framework.FieldSchema {
				"key": &framework.FieldSchema{
					Type: framework.TypeString,
				},
			},
		}

		value, err := parseDurationField("key", d)
		if tc.expectError {
			assert.NotNil(t, err)
			assert.Equal(t, time.Duration(0), value)
		} else {
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedValue, value)
		}
	}
}

func TestPathConfigRead(t *testing.T) {
	var b backend
	testcases := []struct{
		storage logical.Storage
		expectError bool
		expectedData map[string]interface{}
	} {
		{
			storage: GetErrorStorage{},
			expectError: true,
		},
		{
			storage: GetValidStorage{
				Entry: &logical.StorageEntry{
					Key: "config",
					Value: []byte("{\"circleci_token\":\"test-token\",\"base_url\":\"https://test.circleci.com\", \"ttl\":300000000000, \"max_ttl\":900000000000}"),
				},
			},
			expectError: false,
			expectedData: map[string]interface{} {
				"circleci_token": "test-token",
				"base_url": "https://test.circleci.com",
				"ttl": time.Duration(300),
				"max_ttl": time.Duration(900),
			},
		},
	}

	for _, tc := range testcases {
		resp, err := b.pathConfigRead(&logical.Request{Storage: tc.storage}, nil)

		if tc.expectError {
			assert.NotNil(t, err)
			assert.Nil(t, resp)
		} else {
			assert.Nil(t, err)
			assert.NotNil(t, resp)
			for k, v := range tc.expectedData {
				assert.Equal(t, v, resp.Data[k])
			}			
		}
	}
}

func TestPathConfigWrite(t *testing.T) {
	var b backend
	testcases := []struct{
		fieldData *framework.FieldData
		reqStorage logical.Storage
		expectError bool
		expectErrorResponse bool	
	} {
		{ 
			fieldData: &framework.FieldData{
				Raw: map[string]interface{} {
					"circleci_token": "test-token",
					"base_url": "https://bad#m=%",
					"ttl": "5s",
					"max_ttl": "15s",
				},
				Schema: map[string]*framework.FieldSchema{
					"circleci_token": &framework.FieldSchema{
						Type: framework.TypeString,
					},
					"base_url": &framework.FieldSchema{
						Type: framework.TypeString,
					},
					"ttl": &framework.FieldSchema{
						Type: framework.TypeString,
					},
					"max_ttl": &framework.FieldSchema{
						Type: framework.TypeString,
					},
				},
			},
			expectErrorResponse: true,
		},
		{
			fieldData: &framework.FieldData{
				Raw: map[string]interface{} {
					"circleci_token": "test-token",
					"base_url": "https://test.circleci.com",
					"ttl": "5s",
					"max_ttl": "15s",
				},
				Schema: map[string]*framework.FieldSchema{
					"circleci_token": &framework.FieldSchema{
						Type: framework.TypeString,
					},
					"base_url": &framework.FieldSchema{
						Type: framework.TypeString,
					},
					"ttl": &framework.FieldSchema{
						Type: framework.TypeString,
					},
					"max_ttl": &framework.FieldSchema{
						Type: framework.TypeString,
					},
				},
			},
			reqStorage: &PutErrorStorage{},
			expectError: true,
		},
		{
			fieldData: &framework.FieldData{
				Raw: map[string]interface{} {
					"circleci_token": "test-token",
					"base_url": "https://test.circleci.com",
					"ttl": "5s",
					"max_ttl": "15s",
				},
				Schema: map[string]*framework.FieldSchema{
					"circleci_token": &framework.FieldSchema{
						Type: framework.TypeString,
					},
					"base_url": &framework.FieldSchema{
						Type: framework.TypeString,
					},
					"ttl": &framework.FieldSchema{
						Type: framework.TypeString,
					},
					"max_ttl": &framework.FieldSchema{
						Type: framework.TypeString,
					},
				},
			},
			reqStorage: &PutSuccessfulStorage{},
		},
	}

	for _, tc := range testcases {
		resp, err := b.pathConfigWrite(&logical.Request{Storage: tc.reqStorage}, tc.fieldData)
		if tc.expectError {
			assert.NotNil(t, err)
			assert.Nil(t, resp)
		} else if tc.expectErrorResponse {
			assert.Nil(t, err)
			assert.NotNil(t, resp)
			_, ok := resp.Data["error"]
			assert.True(t, ok)
		} else {
			assert.Nil(t, err)
			assert.Nil(t, resp)
		}
	}
}

type PutSuccessfulStorage struct {
	logical.Storage
}

func (s PutSuccessfulStorage) Put(e *logical.StorageEntry) error {
	return nil
}

type PutErrorStorage struct {
	logical.Storage
}

func (s PutErrorStorage) Put(e *logical.StorageEntry) error {
	return errors.New("some error")
}