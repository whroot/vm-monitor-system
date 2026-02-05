package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateRSAKeyPair(t *testing.T) {
	t.Run("GenerateKeyPair", func(t *testing.T) {
		keyPair, err := GenerateRSAKeyPair()

		assert.NoError(t, err)
		assert.NotNil(t, keyPair)
		assert.NotNil(t, keyPair.PublicKey)
		assert.NotNil(t, keyPair.PrivateKey)
	})

	t.Run("KeyPairConsistency", func(t *testing.T) {
		keyPair1, err := GenerateRSAKeyPair()
		assert.NoError(t, err)

		keyPair2, err := GenerateRSAKeyPair()
		assert.NoError(t, err)

		assert.NotEqual(t, keyPair1.PrivateKey, keyPair2.PrivateKey)
	})
}

func TestEncryptDecrypt(t *testing.T) {
	t.Run("EncryptDecryptSuccess", func(t *testing.T) {
		keyPair, err := GenerateRSAKeyPair()
		assert.NoError(t, err)

		originalData := []byte("secret data")
		encrypted, err := EncryptWithPublicKey(keyPair.PublicKey, originalData)
		assert.NoError(t, err)
		assert.NotEqual(t, originalData, encrypted)

		decrypted, err := DecryptWithPrivateKey(keyPair.PrivateKey, encrypted)
		assert.NoError(t, err)
		assert.Equal(t, originalData, decrypted)
	})

	t.Run("EmptyData", func(t *testing.T) {
		keyPair, err := GenerateRSAKeyPair()
		assert.NoError(t, err)

		encrypted, err := EncryptWithPublicKey(keyPair.PublicKey, []byte{})
		assert.NoError(t, err)

		decrypted, err := DecryptWithPrivateKey(keyPair.PrivateKey, encrypted)
		assert.NoError(t, err)
		assert.Empty(t, decrypted)
	})

	t.Run("LongData", func(t *testing.T) {
		keyPair, err := GenerateRSAKeyPair()
		assert.NoError(t, err)

		longData := make([]byte, 200)
		for i := range longData {
			longData[i] = byte(i % 256)
		}

		encrypted, err := EncryptWithPublicKey(keyPair.PublicKey, longData)
		assert.NoError(t, err)

		decrypted, err := DecryptWithPrivateKey(keyPair.PrivateKey, encrypted)
		assert.NoError(t, err)
		assert.Equal(t, longData, decrypted)
	})
}

func TestPublicKeyToBase64(t *testing.T) {
	keyPair, err := GenerateRSAKeyPair()
	assert.NoError(t, err)

	base64Str := PublicKeyToBase64(keyPair.PublicKey)

	assert.NotEmpty(t, base64Str)
}

func TestPrivateKeyToBase64(t *testing.T) {
	keyPair, err := GenerateRSAKeyPair()
	assert.NoError(t, err)

	base64Str := PrivateKeyToBase64(keyPair.PrivateKey)

	assert.NotEmpty(t, base64Str)
}

func TestLoadPublicKeyFromBase64(t *testing.T) {
	keyPair, err := GenerateRSAKeyPair()
	assert.NoError(t, err)

	base64Str := PublicKeyToBase64(keyPair.PublicKey)
	loadedKey, err := LoadPublicKeyFromBase64(base64Str)

	assert.NoError(t, err)
	assert.NotNil(t, loadedKey)
	assert.Equal(t, keyPair.PublicKey.N, loadedKey.N)
	assert.Equal(t, keyPair.PublicKey.E, loadedKey.E)
}

func TestLoadPrivateKeyFromBase64(t *testing.T) {
	keyPair, err := GenerateRSAKeyPair()
	assert.NoError(t, err)

	base64Str := PrivateKeyToBase64(keyPair.PrivateKey)
	loadedKey, err := LoadPrivateKeyFromBase64(base64Str)

	assert.NoError(t, err)
	assert.NotNil(t, loadedKey)
	assert.Equal(t, keyPair.PrivateKey.N, loadedKey.N)
	assert.Equal(t, keyPair.PrivateKey.E, loadedKey.E)
}

func TestLoadPublicKeyFromPEM(t *testing.T) {
	keyPair, err := GenerateRSAKeyPair()
	assert.NoError(t, err)

	tempFile := "/tmp/test_public_key.pem"
	defer os.Remove(tempFile)

	err = SavePublicKeyToPEM(keyPair.PublicKey, tempFile)
	assert.NoError(t, err)

	loadedKey, err := LoadPublicKeyFromPEM(tempFile)
	assert.NoError(t, err)
	assert.NotNil(t, loadedKey)
}

func TestLoadPrivateKeyFromPEM(t *testing.T) {
	keyPair, err := GenerateRSAKeyPair()
	assert.NoError(t, err)

	tempFile := "/tmp/test_private_key.pem"
	defer os.Remove(tempFile)

	err = SavePrivateKeyToPEM(keyPair.PrivateKey, tempFile)
	assert.NoError(t, err)

	loadedKey, err := LoadPrivateKeyFromPEM(tempFile)
	assert.NoError(t, err)
	assert.NotNil(t, loadedKey)
}

func TestSavePublicKeyToPEM(t *testing.T) {
	keyPair, err := GenerateRSAKeyPair()
	assert.NoError(t, err)

	tempFile := "/tmp/test_save_public.pem"
	defer os.Remove(tempFile)

	err = SavePublicKeyToPEM(keyPair.PublicKey, tempFile)
	assert.NoError(t, err)

	_, err = os.Stat(tempFile)
	assert.NoError(t, err)
}

func TestSavePrivateKeyToPEM(t *testing.T) {
	keyPair, err := GenerateRSAKeyPair()
	assert.NoError(t, err)

	tempFile := "/tmp/test_save_private.pem"
	defer os.Remove(tempFile)

	err = SavePrivateKeyToPEM(keyPair.PrivateKey, tempFile)
	assert.NoError(t, err)

	_, err = os.Stat(tempFile)
	assert.NoError(t, err)
}

func TestInvalidBase64(t *testing.T) {
	t.Run("InvalidPublicKeyBase64", func(t *testing.T) {
		loadedKey, err := LoadPublicKeyFromBase64("invalid-base64-string!!!")
		assert.Error(t, err)
		assert.Nil(t, loadedKey)
	})

	t.Run("InvalidPrivateKeyBase64", func(t *testing.T) {
		loadedKey, err := LoadPrivateKeyFromBase64("invalid-base64-string!!!")
		assert.Error(t, err)
		assert.Nil(t, loadedKey)
	})
}

func TestInvalidPEM(t *testing.T) {
	t.Run("NonExistentPEMFile", func(t *testing.T) {
		loadedKey, err := LoadPublicKeyFromPEM("/non/existent/file.pem")
		assert.Error(t, err)
		assert.Nil(t, loadedKey)
	})

	t.Run("InvalidPEMData", func(t *testing.T) {
		tempFile := "/tmp/invalid_pem.pem"
		defer os.Remove(tempFile)

		os.WriteFile(tempFile, []byte("not a valid pem"), 0644)

		loadedKey, err := LoadPublicKeyFromPEM(tempFile)
		assert.Error(t, err)
		assert.Nil(t, loadedKey)
	})
}

func TestRoundTripPEM(t *testing.T) {
	originalPair, err := GenerateRSAKeyPair()
	assert.NoError(t, err)

	tempPubFile := "/tmp/roundtrip_pub.pem"
	tempPrivFile := "/tmp/roundtrip_priv.pem"
	defer os.Remove(tempPubFile)
	defer os.Remove(tempPrivFile)

	err = SavePublicKeyToPEM(originalPair.PublicKey, tempPubFile)
	assert.NoError(t, err)

	err = SavePrivateKeyToPEM(originalPair.PrivateKey, tempPrivFile)
	assert.NoError(t, err)

	loadedPubKey, err := LoadPublicKeyFromPEM(tempPubFile)
	assert.NoError(t, err)

	loadedPrivKey, err := LoadPrivateKeyFromPEM(tempPrivFile)
	assert.NoError(t, err)

	testData := []byte("roundtrip test data")
	encrypted, err := EncryptWithPublicKey(loadedPubKey, testData)
	assert.NoError(t, err)

	decrypted, err := DecryptWithPrivateKey(loadedPrivKey, encrypted)
	assert.NoError(t, err)
	assert.Equal(t, testData, decrypted)
}
