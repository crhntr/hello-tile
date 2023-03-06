package manifest

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/pivotal-cf/planitest"
	"github.com/stretchr/testify/require"
)

type ProductConfiguration = map[string]any

func TestManifest(t *testing.T) {
	t.Setenv("RENDERER", "ops-manifest")

	product, err := planitest.NewProductService(createProductConfig(t))
	require.NoError(t, err)

	t.Run("port", func(t *testing.T) {
		for _, tt := range []struct {
			Name          string
			Config        ProductConfiguration
			ExpectedValue int
		}{
			{Name: "Default", Config: ProductConfiguration{}, ExpectedValue: 8080},
			{Name: "Not Default", Config: ProductConfiguration{".properties.port": 8888}, ExpectedValue: 8888},
		} {
			t.Run(tt.Name, func(t *testing.T) {
				manifest, err := product.RenderManifest(tt.Config)
				require.NoError(t, err)

				helloServerManifest, err := manifest.FindInstanceGroupJob("hello-server", "hello-server")
				require.NoError(t, err)
				value, err := helloServerManifest.Property("port")
				require.NoError(t, err)

				require.Equal(t, value, tt.ExpectedValue)
			})
		}
	})
}

func createProductConfig(t *testing.T) planitest.ProductConfig {
	t.Helper()
	m := generateMetadataFile(t)

	// writes the tile metadata to a file
	// this is required because planitest.ProductConfig expects two io.ReadSeeker fields.
	// ideally it would just take a byte buffer
	tmp := t.TempDir()
	t.Cleanup(func() {
		_ = os.RemoveAll(tmp)
	})
	mp := filepath.Join(tmp, "metadata.yml")
	metadataFile, err := os.Create(mp)
	require.NoError(t, err)
	t.Cleanup(func() {
		closeAndIgnoreError(metadataFile)
	})
	_, err = metadataFile.Write(m)
	require.NoError(t, err)
	t.Cleanup(func() {
		closeAndIgnoreError(metadataFile)
	})

	configFile, err := os.Open("config.yml")
	require.NoError(t, err)
	t.Cleanup(func() {
		closeAndIgnoreError(configFile)
	})

	return planitest.ProductConfig{
		TileFile:   metadataFile,
		ConfigFile: configFile,
	}
}

func generateMetadataFile(t *testing.T) []byte {
	t.Helper()
	// generates tile metadata using kiln
	_, err := exec.LookPath("kiln")
	require.NoError(t, err, "kiln must be installed to run the tests https://github.com/pivotal-cf/kiln")
	wd, err := os.Getwd()
	require.NoError(t, err)
	tileDirectory := filepath.Dir(filepath.Dir(wd))
	bake := exec.Command("kiln", "bake", "--metadata-only", "--stub-releases")
	bake.Dir = tileDirectory
	var out bytes.Buffer
	bake.Stdout = &out
	bake.Stderr = os.Stderr
	bakeErr := bake.Run()
	require.NoError(t, bakeErr, "failed to run kiln bake")
	return out.Bytes()
}

func closeAndIgnoreError(c io.Closer) {
	_ = c.Close()
}
