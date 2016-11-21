package acra

import (
	"github.com/cossacklabs/themis/gothemis/keys"
	"io"
)

type PgDecryptor struct {
	is_with_zone      bool
	key_store         KeyStore
	zone_matcher      *ZoneIdMatcher
	pg_decryptor      DataDecryptor
	binary_decryptor  DataDecryptor
	matched_decryptor DataDecryptor

	poison_key       []byte
	callback_storage *PoisonCallbackStorage
}

func NewPgDecryptor(decryptor DataDecryptor) *PgDecryptor {
	return &PgDecryptor{
		is_with_zone:     false,
		pg_decryptor:     decryptor,
		binary_decryptor: NewBinaryDecryptor(),
	}
}

func (decryptor *PgDecryptor) SetPoisonKey(key []byte) {
	decryptor.poison_key = key
}

func (decryptor *PgDecryptor) GetPoisonKey() []byte {
	return decryptor.poison_key
}

func (decryptor *PgDecryptor) SetWithZone(b bool) {
	decryptor.is_with_zone = b
}

func (decryptor *PgDecryptor) SetZoneMatcher(zone_matcher *ZoneIdMatcher) {
	decryptor.zone_matcher = zone_matcher
}

func (decryptor *PgDecryptor) IsMatchedZone() bool {
	return decryptor.zone_matcher.IsMatched() && decryptor.key_store.HasKey(decryptor.zone_matcher.GetZoneId())
}

func (decryptor *PgDecryptor) MatchZone(b byte) bool {
	return decryptor.zone_matcher.Match(b)
}

func (decryptor *PgDecryptor) GetMatchedZoneId() []byte {
	if decryptor.IsWithZone() {
		return decryptor.zone_matcher.GetZoneId()
	} else {
		return nil
	}
}

func (decryptor *PgDecryptor) ResetZoneMatch() {
	decryptor.zone_matcher.Reset()
}

func (decryptor *PgDecryptor) MatchBeginTag(char byte) bool {
	/* should be called two decryptors */
	matched := decryptor.pg_decryptor.MatchBeginTag(char)
	matched = matched || decryptor.binary_decryptor.MatchBeginTag(char)
	return matched
}

func (decryptor *PgDecryptor) IsWithZone() bool {
	return decryptor.is_with_zone
}

func (decryptor *PgDecryptor) IsMatched() bool {
	if decryptor.binary_decryptor.IsMatched() {
		decryptor.matched_decryptor = decryptor.binary_decryptor
		return true
	} else if decryptor.pg_decryptor.IsMatched() {
		decryptor.matched_decryptor = decryptor.pg_decryptor
		return true
	} else {
		decryptor.matched_decryptor = nil
		return false
	}
}
func (decryptor *PgDecryptor) Reset() {
	decryptor.matched_decryptor = nil
	decryptor.binary_decryptor.Reset()
	decryptor.pg_decryptor.Reset()
}
func (decryptor *PgDecryptor) GetMatched() []byte {
	if decryptor.matched_decryptor != nil {
		return decryptor.matched_decryptor.GetMatched()
	} else {
		return decryptor.pg_decryptor.GetMatched()
	}
}
func (decryptor *PgDecryptor) ReadSymmetricKey(private_key *keys.PrivateKey, reader io.Reader) ([]byte, []byte, error) {
	return decryptor.matched_decryptor.ReadSymmetricKey(private_key, reader)
}

func (decryptor *PgDecryptor) ReadData(symmetric_key, zone_id []byte, reader io.Reader) ([]byte, error) {
	return decryptor.matched_decryptor.ReadData(symmetric_key, zone_id, reader)
}

func (decryptor *PgDecryptor) SetKeyStore(store KeyStore) {
	decryptor.key_store = store
}

func (decryptor *PgDecryptor) GetPrivateKey() (*keys.PrivateKey, error) {
	return decryptor.key_store.GetKey(decryptor.GetMatchedZoneId())
}

func (decryptor *PgDecryptor) GetPoisonCallbackStorage() *PoisonCallbackStorage {
	return decryptor.callback_storage
}

func (decryptor *PgDecryptor) SetPoisonCallbackStorage(storage *PoisonCallbackStorage) {
	decryptor.callback_storage = storage
}