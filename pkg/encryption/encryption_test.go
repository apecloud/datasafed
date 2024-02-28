package encryption_test

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/apecloud/datasafed/pkg/encryption"
	"github.com/stretchr/testify/require"
)

func TestRoundTrip(t *testing.T) {
	data := make([]byte, 100)
	rand.Read(data)

	passPhrase := make([]byte, 32)
	rand.Read(passPhrase)

	for _, encryptionAlgo := range encryption.SupportedAlgorithms() {
		encryptionAlgo := encryptionAlgo
		t.Run(encryptionAlgo, func(t *testing.T) {
			e, err := encryption.CreateEncryptor(encryptionAlgo, passPhrase)
			if err != nil {
				t.Fatal(err)
			}

			var cipherText1 bytes.Buffer
			var cipherText2 bytes.Buffer

			require.NoError(t, e.EncryptStream(bytes.NewBuffer(data), &cipherText1))
			require.NoError(t, e.EncryptStream(bytes.NewBuffer(data), &cipherText2))

			if v := cipherText1.Bytes(); bytes.Equal(v, cipherText2.Bytes()) {
				t.Errorf("multiple EncryptStream returned the same ciphertext: %x", v)
			}

			var plainText1 bytes.Buffer

			require.NoError(t, e.DecryptStream(&cipherText1, &plainText1))

			if v := plainText1.Bytes(); !bytes.Equal(v, data) {
				t.Errorf("EncryptStream()/DecryptStream() does not round-trip: %x %x", v, data)
			}

			var plainText2 bytes.Buffer

			require.NoError(t, e.DecryptStream(bytes.NewBuffer(cipherText2.Bytes()), &plainText2))

			if v := plainText2.Bytes(); !bytes.Equal(v, data) {
				t.Errorf("EncryptStream()/DecryptStream() does not round-trip: %x %x", v, data)
			}

			// TODO: enable the following logic if the algorithm is AEAD
			// // flip some bits in the cipherText
			// b := cipherText2.Bytes()
			// b[mathrand.Intn(len(b))] ^= byte(1 + mathrand.Intn(254))

			// plainText2.Reset()
			// require.Error(t, e.DecryptStream(bytes.NewBuffer(cipherText2.Bytes()), &plainText2))
		})
	}
}

func TestCiphertextSamples(t *testing.T) {
	cases := []struct {
		passPhrase []byte
		payload    []byte
		samples    map[string]string
	}{
		{
			passPhrase: []byte("01234567890123456789012345678901"), // 32 bytes
			payload:    []byte("foo"),

			// samples of base16-encoded ciphertexts of payload encrypted with passPhrase
			samples: map[string]string{
				"AES128-CFB": "3f531b215a8b0774edeb5f07f451f811c6ba0b",
				"AES192-CFB": "65cce058982cde6dec94ee7965c737bd2e9044",
				"AES256-CFB": "cd68a8ba7e886f00326ebd9da560bfca0ad5c4",
			},
		},
		{
			passPhrase: []byte("abcdefghijklmnopqrstuvwxyzabcdef"), // 32 bytes
			payload:    []byte("quick brown fox jumps over the lazy dog"),

			// samples of base16-encoded ciphertexts of payload encrypted with passPhrase
			samples: map[string]string{
				"AES128-CFB": "a4fd5ed9b98b780f09c3253dadd81e9b96b52f3fbe215ab0b43e88df82457b5eb4209bdabe4d8edf045763d17807ea559f4d1e316edc9b",
				"AES192-CFB": "6d9cfd25a3b5f2299534c87cd6f61f16c153178e74496d9b6b67f351d9c2a4a7c1514a5a4b42efe945ac56baea71f1dff51df9dc40a8a4",
				"AES256-CFB": "9456421484adb715d7f6b52663908dd1acf16848077df01942847cc0e835a627c8b5c704b465ea86f47afd4e359c097582e81a544fdbd1",
			},
		},
	}

	for _, tc := range cases {
		verifyCiphertextSamples(t, tc.passPhrase, tc.payload, tc.samples)
	}
}

func verifyCiphertextSamples(t *testing.T, passPhrase, payload []byte, samples map[string]string) {
	t.Helper()

	for _, encryptionAlgo := range encryption.SupportedAlgorithms() {
		enc, err := encryption.CreateEncryptor(encryptionAlgo, passPhrase)
		if err != nil {
			t.Fatal(err)
		}

		ct := samples[encryptionAlgo]
		if ct == "" {
			func() {
				var v bytes.Buffer
				require.NoError(t, enc.EncryptStream(bytes.NewBuffer(payload), &v))

				t.Errorf("missing ciphertext sample for %q: %q, possible one is: %q",
					encryptionAlgo, payload, hex.EncodeToString(v.Bytes()))
			}()
		} else {
			b, err := hex.DecodeString(ct)
			if err != nil {
				t.Errorf("invalid ciphertext for %v: %v", encryptionAlgo, err)
				continue
			}

			func() {
				var plainText bytes.Buffer

				require.NoError(t, enc.DecryptStream(bytes.NewBuffer(b), &plainText))

				if v := plainText.Bytes(); !bytes.Equal(v, payload) {
					t.Errorf("invalid plaintext after decryption %x, want %x", v, payload)
				}
			}()
		}
	}
}

func benchmarkEncryption(b *testing.B, algorithm string) {
	passPhrase := make([]byte, 32)
	rand.Read(passPhrase)

	enc, err := encryption.CreateEncryptor(algorithm, passPhrase)
	require.NoError(b, err)

	// 8 MiB
	plainText := bytes.Repeat([]byte{1, 2, 3, 4, 5, 6, 7, 8}, 1<<20)

	var warmupOut bytes.Buffer
	require.NoError(b, enc.EncryptStream(bytes.NewBuffer(plainText), &warmupOut))

	b.ResetTimer()

	var out bytes.Buffer
	out.Grow(len(plainText) + enc.Overhead())
	for i := 0; i < b.N; i++ {
		out.Reset()
		enc.EncryptStream(bytes.NewBuffer(plainText), &out)
		b.SetBytes(int64(len(plainText)))
	}
}

func BenchmarkEncryption(b *testing.B) {
	for _, encryptionAlgo := range encryption.SupportedAlgorithms() {
		encryptionAlgo := encryptionAlgo
		b.Run(encryptionAlgo, func(b *testing.B) {
			benchmarkEncryption(b, encryptionAlgo)
		})
	}
}

func benchmarkDecryption(b *testing.B, algorithm string) {
	passPhrase := make([]byte, 32)
	rand.Read(passPhrase)

	enc, err := encryption.CreateEncryptor(algorithm, passPhrase)
	require.NoError(b, err)

	// 8 MiB
	plainText := bytes.Repeat([]byte{1, 2, 3, 4, 5, 6, 7, 8}, 1<<20)

	var warmupOut bytes.Buffer
	require.NoError(b, enc.EncryptStream(bytes.NewBuffer(plainText), &warmupOut))

	cipherText := warmupOut.Bytes()

	b.ResetTimer()

	var out bytes.Buffer
	out.Grow(len(plainText))
	for i := 0; i < b.N; i++ {
		out.Reset()
		enc.DecryptStream(bytes.NewBuffer(cipherText), &out)
		b.SetBytes(int64(len(cipherText)))
	}
}

func BenchmarkDecryption(b *testing.B) {
	for _, encryptionAlgo := range encryption.SupportedAlgorithms() {
		encryptionAlgo := encryptionAlgo
		b.Run(encryptionAlgo, func(b *testing.B) {
			benchmarkDecryption(b, encryptionAlgo)
		})
	}
}
