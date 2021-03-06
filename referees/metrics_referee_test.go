package referees

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/common/model"

	"github.com/prometheus/client_golang/api"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewMetricsRefereeNoConfig(t *testing.T) {
	mockExecutor := new(mockMetricsExecutor)
	config := &Config{}
	log := logrus.WithField("test", 1)
	mr := CreateMetricsReferee(mockExecutor, config, log)
	require.Nil(t, mr)
}

func TestNewMetricsRefereeImproperExecutor(t *testing.T) {
	mockExecutor := struct{}{}
	config := &Config{
		Metrics: &MetricsRefereeConfig{
			PrometheusAddress: "http://localhost:9000",
			QueryInterval:     10,
			Queries:           []string{"name1:metric1{{selector}}", "name2:metric2{{selector}}"},
		},
	}

	log := logrus.WithField("test", 1)
	mr := CreateMetricsReferee(mockExecutor, config, log)
	require.Nil(t, mr)
}

func TestNewMetricsRefereeBadPrometheusAddress(t *testing.T) {
	mockExecutor := new(mockMetricsExecutor)
	config := &Config{
		Metrics: &MetricsRefereeConfig{
			PrometheusAddress: "*(^&*^*(34f34f34fg3rfg3rgfY&*^^%*&^*(^(*",
			QueryInterval:     10,
			Queries:           []string{"name1:metric1{{selector}}", "name2:metric2{{selector}}"},
		},
	}

	log := logrus.WithField("test", 1)
	mr := CreateMetricsReferee(mockExecutor, config, log)
	require.Nil(t, mr)
}

func TestNewMetricsReferee(t *testing.T) {
	mockExecutor := new(mockMetricsExecutor)
	config := &Config{
		Metrics: &MetricsRefereeConfig{
			PrometheusAddress: "http://localhost:9000",
			QueryInterval:     10,
			Queries:           []string{"name1:metric1{{selector}}", "name2:metric2{{selector}}"},
		},
	}

	log := logrus.WithField("test", 1)
	mr := CreateMetricsReferee(mockExecutor, config, log)
	require.NotNil(t, mr)

	// test job artifact parameters
	assert.Equal(t, "metrics_referee.json", mr.ArtifactBaseName())
	assert.Equal(t, "metrics_referee", mr.ArtifactType())
	assert.Equal(t, "gzip", mr.ArtifactFormat())
}

func TestMetricsRefereeExecuteParseError(t *testing.T) {
	mockExecutor := new(mockMetricsExecutor)
	config := &Config{
		Metrics: &MetricsRefereeConfig{
			PrometheusAddress: "http://localhost:9000",
			QueryInterval:     10,
			Queries:           []string{"name1=metric1{{selector}}", "name2=metric2{{selector}}"},
		},
	}

	log := logrus.WithField("test", 1)
	mr := CreateMetricsReferee(mockExecutor, config, log)
	require.NotNil(t, mr)

	ctx := context.Background()
	_, err := mr.Execute(ctx, time.Now(), time.Now())
	require.Error(t, err)
}

func TestMetricsRefereeExecuteQueryRangeError(t *testing.T) {
	mockExecutor := new(mockMetricsExecutor)
	config := &Config{
		Metrics: &MetricsRefereeConfig{
			PrometheusAddress: "http://localhost:9000",
			QueryInterval:     10,
			Queries:           []string{"name1:metric1{{selector}}", "name2:metric2{{selector}}"},
		},
	}

	log := logrus.WithField("test", 1)
	mr := CreateMetricsReferee(mockExecutor, config, log)
	require.NotNil(t, mr)

	ctx := context.Background()
	prometheusAPI := new(MockPrometheusAPI)
	matrix := model.Matrix([]*model.SampleStream{})
	prometheusAPI.On("QueryRange", mock.Anything, mock.Anything, mock.Anything).Return(matrix, api.Warnings([]string{}), errors.New("test"))

	mr.prometheusAPI = prometheusAPI
	_, err := mr.Execute(ctx, time.Now(), time.Now())
	require.NoError(t, err)
}

func TestMetricsRefereeExecuteQueryRangeNonMatrixReturn(t *testing.T) {
	mockExecutor := new(mockMetricsExecutor)
	config := &Config{
		Metrics: &MetricsRefereeConfig{
			PrometheusAddress: "http://localhost:9000",
			QueryInterval:     10,
			Queries:           []string{"name1:metric1{{selector}}", "name2:metric2{{selector}}"},
		},
	}

	log := logrus.WithField("test", 1)
	mr := CreateMetricsReferee(mockExecutor, config, log)
	require.NotNil(t, mr)

	ctx := context.Background()
	prometheusAPI := new(MockPrometheusAPI)
	prometheusAPI.On("QueryRange", mock.Anything, mock.Anything, mock.Anything).Return(&MockPrometheusValue{}, api.Warnings([]string{}), nil)

	mr.prometheusAPI = prometheusAPI
	_, err := mr.Execute(ctx, time.Now(), time.Now())
	require.NoError(t, err)
}

func TestMetricsRefereeExecuteQueryRangeResultEmpty(t *testing.T) {
	mockExecutor := new(mockMetricsExecutor)
	config := &Config{
		Metrics: &MetricsRefereeConfig{
			PrometheusAddress: "http://localhost:9000",
			QueryInterval:     10,
			Queries:           []string{"name1:metric1{{selector}}", "name2:metric2{{selector}}"},
		},
	}

	log := logrus.WithField("test", 1)
	mr := CreateMetricsReferee(mockExecutor, config, log)
	require.NotNil(t, mr)

	matrix := model.Matrix([]*model.SampleStream{})
	ctx := context.Background()
	prometheusAPI := new(MockPrometheusAPI)
	prometheusAPI.On("QueryRange", mock.Anything, mock.Anything, mock.Anything).Return(matrix, api.Warnings([]string{}), nil)

	mr.prometheusAPI = prometheusAPI
	_, err := mr.Execute(ctx, time.Now(), time.Now())
	require.NoError(t, err)
}

func TestMetricsRefereeExecute(t *testing.T) {
	startTime := time.Unix(1405544146, 0)
	endTime := time.Unix(1405544246, 0)
	response := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"resultType": "matrix",
			"result": []interface{}{
				map[string]interface{}{
					"metric": map[string]string{
						"__name__": "metric1",
						"job":      "prometheus",
						"instance": "localhost:9090",
					},
					"values": []interface{}{
						[]interface{}{1435781430.781, "1"},
					},
				},
				map[string]interface{}{
					"metric": map[string]string{
						"__name__": "metric2",
						"job":      "prometheus",
						"instance": "localhost:9090",
					},
					"values": []interface{}{
						[]interface{}{1435781430.781, "1"},
					},
				},
			},
		},
	}
	responseJSON, err := json.Marshal(response)
	require.NoError(t, err)

	requestIndex := 1
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// parse request
		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(r.Body)
		require.NoError(t, err)
		actual := buf.String()
		t.Log("REQUEST: " + actual)
		query := fmt.Sprintf("metric%d", requestIndex)
		expected := fmt.Sprintf("end=%d&query=%s%%7Bname%%3D%%22value%%22%%7D&start=%d&step=10", endTime.Unix(), query, startTime.Unix())
		// validate request
		require.Equal(t, expected, actual)
		// send response
		t.Log("RESPONSE: " + string(responseJSON))
		_, err = w.Write(responseJSON)
		require.NoError(t, err)
		requestIndex++
	}))
	defer ts.Close()

	mockExecutor := new(mockMetricsExecutor)
	config := &Config{
		Metrics: &MetricsRefereeConfig{
			PrometheusAddress: ts.URL,
			QueryInterval:     10,
			Queries:           []string{"name1:metric1{{selector}}", "name2:metric2{{selector}}"},
		},
	}

	log := logrus.WithField("test", t.Name())
	mr := CreateMetricsReferee(mockExecutor, config, log)
	require.NotNil(t, mr)

	ctx := context.Background()
	reader, err := mr.Execute(ctx, startTime, endTime)
	require.NoError(t, err)

	// convert reader result to golang maps
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(reader)
	require.NoError(t, err)
	var metrics interface{}
	err = json.Unmarshal(buf.Bytes(), &metrics)
	require.NoError(t, err)

	// confirm length of elements
	assert.Len(t, metrics, len(config.Metrics.Queries))
}
