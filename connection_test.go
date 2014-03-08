package libvirt

import (
	"io/ioutil"
	"testing"
)

const (
	QEMUSystemURI  = "qemu:///system"
	TestDefaultURI = "test:///default"
)

func openTestConnection(t testing.TB) Connection {
	conn, err := Open(QEMUSystemURI)
	if err != nil {
		t.Fatal(err)
	}

	return conn
}

func TestOpen(t *testing.T) {
	if _, err := Open("xxx"); err == nil {
		t.Error("an error was not returned when connecting to a bad URI")
	}

	conn, err := Open(QEMUSystemURI)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	if !conn.IsAlive() {
		t.Error("the libvirt connection was opened but it is not alive")
	}

	// IsEncrypted

	if !conn.IsSecure() {
		t.Error("the libvirt connection is not secure")
	}
}

func TestOpenReadOnly(t *testing.T) {
	if _, err := OpenReadOnly("xxx"); err == nil {
		t.Error("an error was not returned when connecting (RO) to a bad URI")
	}

	conn, err := OpenReadOnly(QEMUSystemURI)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	if !conn.IsAlive() {
		t.Error("the libvirt connection was opened but it is not alive")
	}

	// IsEncrypted

	if !conn.IsSecure() {
		t.Error("the libvirt connection is not secure")
	}
}

func TestVersion(t *testing.T) {
	conn := openTestConnection(t)
	defer conn.Close()

	version, err := conn.Version()
	if err != nil {
		t.Fatal(err)
	}

	if version < 0 {
		t.Errorf("hypervisor version should be a positive number: %d", version)
	}
}

func TestLibVersion(t *testing.T) {
	conn := openTestConnection(t)
	defer conn.Close()

	version, err := conn.LibVersion()
	if err != nil {
		t.Fatal(err)
	}

	if version < 0 {
		t.Errorf("libvirt version should be a positive number: %d", version)
	}
}

func TestCapabilities(t *testing.T) {
	conn := openTestConnection(t)
	defer conn.Close()

	cap, err := conn.Capabilities()
	if err != nil {
		t.Fatal(err)
	}

	if len(cap) == 0 {
		t.Error("libvirt capabilities should not be empty")
	}
}

func TestHostname(t *testing.T) {
	conn := openTestConnection(t)
	defer conn.Close()

	hostname, err := conn.Hostname()
	if err != nil {
		t.Fatal(err)
	}

	if len(hostname) == 0 {
		t.Error("libvirt hostname should not be empty")
	}
}

func TestSysinfo(t *testing.T) {
	conn := openTestConnection(t)
	defer conn.Close()

	sysinfo, err := conn.Sysinfo()
	if err != nil {
		t.Fatal(err)
	}

	if len(sysinfo) == 0 {
		t.Error("libvirt sysinfo should not be empty")
	}
}

func TestType(t *testing.T) {
	conn := openTestConnection(t)
	defer conn.Close()

	typ, err := conn.Type()
	if err != nil {
		t.Fatal(err)
	}

	if len(typ) == 0 {
		t.Error("libvirt type should not be empty")
	}
}

func TestURI(t *testing.T) {
	conn := openTestConnection(t)
	defer conn.Close()

	uri, err := conn.URI()
	if err != nil {
		t.Fatal(err)
	}

	if uri != QEMUSystemURI {
		t.Errorf("libvirt URI should be the same used to open the connection; got=%s, want=%s", uri, QEMUSystemURI)
	}
}

func TestRef(t *testing.T) {
	conn := openTestConnection(t)

	if err := conn.Ref(); err != nil {
		t.Fatal(err)
	}

	if _, err := conn.Close(); err != nil {
		t.Error(err)
	}
	if _, err := conn.Close(); err != nil {
		t.Error("could not close the connection for the second time after calling Ref")
	}
}

func TestCPUModelNames(t *testing.T) {
	conn := openTestConnection(t)
	defer conn.Close()

	if _, err := conn.CPUModelNames("xxx"); err == nil {
		t.Error("an error was not returned when getting CPU model names from invalid arch")
	}

	models, err := conn.CPUModelNames("x86_64")
	if err != nil {
		t.Fatal(err)
	}

	if len(models) == 0 {
		t.Error("libvirt CPU model names should not be empty")
	}
}

func TestMaxVCPUs(t *testing.T) {
	conn := openTestConnection(t)
	defer conn.Close()

	if _, err := conn.MaxVCPUs("xxx"); err == nil {
		t.Error("an error was not returned when getting maximum VCPUs from invalid type")
	}

	vcpus, err := conn.MaxVCPUs("kvm")
	if err != nil {
		t.Fatal(err)
	}

	if vcpus < 0 {
		t.Error("libvirt maximum VCPU should be a positive number")
	}
}

func TestListDomains(t *testing.T) {
	conn := openTestConnection(t)
	defer conn.Close()

	domains, err := conn.ListDomains(DomAll)
	if err != nil {
		t.Fatal(err)
	}

	for _, d := range domains {
		if err := d.Free(); err != nil {
			t.Error(err)
		}
	}
}

func TestCreateAndDestroyDomain(t *testing.T) {
	conn := openTestConnection(t)
	defer conn.Close()

	if _, err := conn.CreateDomain("", DomCreateDefault); err == nil {
		t.Error("an error was not returned when creating a domain with empty XML description")
	}

	xml, ioerr := ioutil.ReadFile(DomTestXMLFile)
	if ioerr != nil {
		t.Fatal(ioerr)
	}

	if _, err := conn.CreateDomain(string(xml), DomainCreateFlag(99)); err == nil {
		t.Error("an error was not returned when using an invalid create flag")
	}

	dom, err := conn.CreateDomain(string(xml), DomCreateDefault)
	if err != nil {
		t.Fatal(err)
	}
	defer dom.Free()

	if err := dom.Destroy(DomDestroyDefault); err != nil {
		t.Error(err)
	}
}

func TestDefineAndUndefineDomain(t *testing.T) {
	conn := openTestConnection(t)
	defer conn.Close()

	if _, err := conn.DefineDomain(""); err == nil {
		t.Error("an error was not returned when defining a domain with empty XML description")
	}

	xml, ioerr := ioutil.ReadFile(DomTestXMLFile)
	if ioerr != nil {
		t.Fatal(ioerr)
	}

	dom, err := conn.DefineDomain(string(xml))
	if err != nil {
		t.Fatal(err)
	}
	defer dom.Free()

	if err := dom.Undefine(DomainUndefineFlag(99)); err == nil {
		t.Error("an error was not return when using an invalid undefine flag")
	}

	if err := dom.Undefine(DomUndefineDefault); err != nil {
		t.Error(err)
	}
}

func TestLookupDomainByID(t *testing.T) {
	dom, conn := createTestDomain(t, DomCreateStartAutodestroy)
	defer conn.Close()
	defer dom.Free()

	if _, err := conn.LookupDomainByID(99); err == nil {
		t.Error("an error was not returned when looking up a non-existing domain ID")
	}

	expectedID, err := dom.ID()
	if err != nil {
		t.Fatal(err)
	}

	dom, err = conn.LookupDomainByID(expectedID)
	if err != nil {
		t.Fatal(err)
	}
	defer dom.Free()

	id, err := dom.ID()
	if err != nil {
		t.Error(err)
	}

	if id != expectedID {
		t.Errorf("looked up domain with unexpected id; got=%d, want=%d", id, expectedID)
	}
}

func TestLookupDomainByName(t *testing.T) {
	dom, conn := createTestDomain(t, DomCreateStartAutodestroy)
	defer conn.Close()
	defer dom.Free()

	if _, err := conn.LookupDomainByName("xxx"); err == nil {
		t.Error("an error was not returned when looking up a non-existing domain name")
	}

	dom, err := conn.LookupDomainByName(DomTestName)
	if err != nil {
		t.Fatal(err)
	}
	defer dom.Free()

	if name := dom.Name(); name != DomTestName {
		t.Errorf("looked up domain with unexpected name; got=%s, want=%s", name, DomTestName)
	}
}

func TestLookupDomainByUUID(t *testing.T) {
	dom, conn := createTestDomain(t, DomCreateStartAutodestroy)
	defer conn.Close()
	defer dom.Free()

	if _, err := conn.LookupDomainByUUID("xxx"); err == nil {
		t.Error("an error was not returned when looking up a non-existing domain UUID")
	}

	dom, err := conn.LookupDomainByUUID(DomTestUUID)
	if err != nil {
		t.Fatal(err)
	}
	defer dom.Free()

	uuid, err := dom.UUID()
	if err != nil {
		t.Error(err)
	}

	if uuid != DomTestUUID {
		t.Errorf("looked up domain with unexpected UUID; got=%s, want=%s", uuid, DomTestUUID)
	}
}

func BenchmarkConnection(b *testing.B) {
	for n := 0; n < b.N; n++ {
		conn, err := Open(QEMUSystemURI)
		if err != nil {
			b.Error(err)
		}

		if _, err := conn.Close(); err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkTestConnection(b *testing.B) {
	for n := 0; n < b.N; n++ {
		conn, err := Open(TestDefaultURI)
		if err != nil {
			b.Error(err)
		}

		if _, err := conn.Close(); err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkCreateDomain(b *testing.B) {
	conn, err := Open(QEMUSystemURI)
	if err != nil {
		b.Fatal(err)
	}

	xml, ioerr := ioutil.ReadFile(DomTestXMLFile)
	if ioerr != nil {
		b.Fatal(ioerr)
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		dom, err := conn.CreateDomain(string(xml), DomCreateDefault)
		if err != nil {
			b.Error(err)
		}

		if err := dom.Destroy(DomDestroyDefault); err != nil {
			b.Error(err)
		}

		if err := dom.Free(); err != nil {
			b.Error(err)
		}
	}
	b.StopTimer()

	if _, err := conn.Close(); err != nil {
		b.Error(err)
	}
}

func BenchmarkDefineDomain(b *testing.B) {
	conn, err := Open(QEMUSystemURI)
	if err != nil {
		b.Fatal(err)
	}

	xml, ioerr := ioutil.ReadFile(DomTestXMLFile)
	if ioerr != nil {
		b.Fatal(ioerr)
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		dom, err := conn.DefineDomain(string(xml))
		if err != nil {
			b.Error(err)
		}

		if err := dom.Undefine(DomUndefineDefault); err != nil {
			b.Error(err)
		}

		if err := dom.Free(); err != nil {
			b.Error(err)
		}
	}
	b.StopTimer()

	if _, err := conn.Close(); err != nil {
		b.Error(err)
	}
}