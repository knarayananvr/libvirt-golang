package libvirt

// #include <stdlib.h>
// #include <libvirt/libvirt.h>
import "C"
import (
	"log"
	"unicode/utf8"
	"unsafe"
)

// SecretListFlag defines a filter when listing secrets.
type SecretListFlag uint32

// Possible values for SecretListFlag.
const (
	SecListAll         SecretListFlag = 0
	SecListEphemeral   SecretListFlag = C.VIR_CONNECT_LIST_SECRETS_EPHEMERAL
	SecListNoEphemeral SecretListFlag = C.VIR_CONNECT_LIST_SECRETS_NO_EPHEMERAL
	SecListPrivate     SecretListFlag = C.VIR_CONNECT_LIST_SECRETS_PRIVATE
	SecListNoPrivate   SecretListFlag = C.VIR_CONNECT_LIST_SECRETS_NO_PRIVATE
)

// SecretUsageType defines a type of secret.
type SecretUsageType uint32

// Possible values for SecretUsageType.
const (
	SecUsageTypeNone   SecretUsageType = C.VIR_SECRET_USAGE_TYPE_NONE
	SecUsageTypeVolume SecretUsageType = C.VIR_SECRET_USAGE_TYPE_VOLUME
	SecUsageTypeCeph   SecretUsageType = C.VIR_SECRET_USAGE_TYPE_CEPH
	SecUsageTypeISCSI  SecretUsageType = C.VIR_SECRET_USAGE_TYPE_ISCSI
)

// Secret holds a libvirt secret. There are no exported fields.
type Secret struct {
	log       *log.Logger
	virSecret C.virSecretPtr
}

// Free releases the secret handle. The underlying secret continues to exist.
func (sec Secret) Free() error {
	sec.log.Println("freeing secret...")
	cRet := C.virSecretFree(sec.virSecret)
	ret := int(cRet)

	if ret == -1 {
		err := LastError()
		sec.log.Printf("an error occurred: %v\n", err)
		return err
	}

	sec.log.Println("secret freed")

	return nil
}

// Undefine deletes the specified secret. This does not free the associated
// "Secret" object.
func (sec Secret) Undefine() error {
	sec.log.Println("undefining secret...")
	cRet := C.virSecretUndefine(sec.virSecret)
	ret := int(cRet)

	if ret == -1 {
		err := LastError()
		sec.log.Printf("an error occurred: %v\n", err)
		return err
	}

	sec.log.Println("secret undefined")

	return nil
}

// UUID fetches the UUID of the secret.
func (sec Secret) UUID() (string, error) {
	cUUID := (*C.char)(C.malloc(C.size_t(C.VIR_UUID_STRING_BUFLEN)))
	defer C.free(unsafe.Pointer(cUUID))

	sec.log.Println("reading secret UUID...")
	cRet := C.virSecretGetUUIDString(sec.virSecret, cUUID)
	ret := int32(cRet)

	if ret == -1 {
		err := LastError()
		sec.log.Printf("an error occurred: %v\n", err)
		return "", err
	}

	uuid := C.GoString(cUUID)
	sec.log.Printf("UUID: %v\n", uuid)

	return uuid, nil
}

// XML fetches an XML document describing attributes of the secret.
func (sec Secret) XML() (string, error) {
	sec.log.Println("reading secret XML...")
	cXML := C.virSecretGetXMLDesc(sec.virSecret, 0)

	if cXML == nil {
		err := LastError()
		sec.log.Printf("an error occurred: %v\n", err)
		return "", err
	}
	defer C.free(unsafe.Pointer(cXML))

	xml := C.GoString(cXML)

	sec.log.Printf("XML length: %v runes\n", utf8.RuneCountInString(xml))

	return xml, nil
}

// UsageID gets the unique identifier of the object with which this secret is to
// be used. The format of the identifier is dependant on the usage type of the
// secret. For a secret with a usage type of SecUsageTypeVolume the identifier
// will be a fully qualfied path name. The identifiers are intended to be unique
// within the set of all secrets sharing the same usage type. ie, there shall
// only ever be one secret for each volume path.
func (sec Secret) UsageID() (string, error) {
	sec.log.Println("reading secret usage ID...")
	cUsageID := C.virSecretGetUsageID(sec.virSecret)

	if cUsageID == nil {
		err := LastError()
		sec.log.Printf("an error occurred: %v\n", err)
		return "", err
	}

	usageID := C.GoString(cUsageID)

	sec.log.Printf("usage ID: %v\n", usageID)

	return usageID, nil
}

// UsageType gets the type of object which uses this secret. The returned value
// is one of the constants defined in the SecretUsageType enumeration. More
// values may be added to this enumeration in the future, so callers should
// expect to see usage types they do not explicitly know about.
func (sec Secret) UsageType() (SecretUsageType, error) {
	sec.log.Println("reading secret usage type...")
	cUsageType := C.virSecretGetUsageType(sec.virSecret)

	if cUsageType == -1 {
		err := LastError()
		sec.log.Printf("an error occurred: %v\n", err)
		return 0, err
	}

	usageType := SecretUsageType(cUsageType)

	sec.log.Printf("usage type: %v\n", usageType)

	return usageType, nil
}

// SetValue sets the value of a secret.
func (sec Secret) SetValue(value string) error {
	cSize := C.size_t(len(value))
	cValue := (*C.uchar)(unsafe.Pointer(C.CString(value)))

	sec.log.Printf("setting secret value (%v)...\n", value)
	cRet := C.virSecretSetValue(sec.virSecret, cValue, cSize, 0)
	ret := int32(cRet)

	if ret == -1 {
		err := LastError()
		sec.log.Printf("an error occurred: %v\n", err)
		return err
	}

	sec.log.Println("value set")

	return nil
}

// Value fetches the value of a secret.
func (sec Secret) Value() (string, error) {
	var cSize C.size_t

	sec.log.Println("reading secret value...")
	cValue := C.virSecretGetValue(sec.virSecret, &cSize, 0)

	if cValue == nil {
		err := LastError()
		sec.log.Printf("an error occurred: %v\n", err)
		return "", err
	}
	defer C.free(unsafe.Pointer(cValue))

	value := string(C.GoBytes(unsafe.Pointer(cValue), C.int(cSize)))
	sec.log.Printf("value: %v\n", value)

	return value, nil
}

// Ref increments the reference count on the secret. For each additional call to
// this method, there shall be a corresponding call to "<Secret>.Free" to
// release the reference count, once the caller no longer needs the reference to
// this object.
// This method is typically useful for applications where multiple threads are
// using a connection, and it is required that the connection remain open until
// all threads have finished using it. ie, each new thread using a secret would
// increment the reference count.
func (sec Secret) Ref() error {
	sec.log.Println("incrementing secret's reference count...")
	cRet := C.virSecretRef(sec.virSecret)
	ret := int32(cRet)

	if ret == -1 {
		err := LastError()
		sec.log.Printf("an error occurred: %v\n", err)
		return err
	}

	sec.log.Println("reference count incremented")

	return nil
}
