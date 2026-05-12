package pubsub_test

import (
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/stackitcloud/stackit-sdk-go/core/auth"
	"github.com/stackitcloud/stackit-sdk-go/core/config"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	environment    string
	topicId        uuid.UUID
	subscriptionId uuid.UUID
	spec           Specification
	rt             http.RoundTripper
)

func TestPubsub(t *testing.T) {
	RegisterFailHandler(Fail)
	SetDefaultEventuallyTimeout(15 * time.Second)
	RunSpecs(t, "PubSub Suite")
}

type Specification struct {
	TopicId             uuid.UUID `envconfig:"TOPIC_ID"              required:"true"`
	SubscriptionId      uuid.UUID `envconfig:"SUBSCRIPTION_ID"       required:"true"`
	ServiceAccountToken string    `envconfig:"SERVICE_ACCOUNT_TOKEN" required:"true"`
	Environment         string    `envconfig:"ENVIRONMENT"           required:"true"`
	TokenCustomUri      string    `envconfig:"TOKEN_CUSTOM_URI"      required:"true"`
}

func findProjectRoot() (string, error) {
	_, b, _, _ := runtime.Caller(0) //nolint:dogsled
	dir := filepath.Dir(b)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("go.mod not found")
		}
		dir = parent
	}
}

var _ = BeforeSuite(func() {
	projectRoot, err := findProjectRoot()
	if err != nil {
		log.Printf("could not find project root: %v", err)
	} else {
		envPath := filepath.Join(projectRoot, ".env")
		if err := godotenv.Load(envPath); err != nil {
			log.Printf("could not load .env file from project root: %v", err)
		}
	}

	err = envconfig.Process("", &spec)
	Expect(err).ToNot(HaveOccurred())

	topicId = spec.TopicId
	subscriptionId = spec.SubscriptionId
	environment = spec.Environment

	rt, err = auth.DefaultAuth(&config.Configuration{
		ServiceAccountKey: spec.ServiceAccountToken,
		TokenCustomUrl:    spec.TokenCustomUri,
	})
	Expect(err).ToNot(HaveOccurred())
})
