module github.com/okdaichi/gomoqt

go 1.25.0

require (
	github.com/magefile/mage v1.15.0
	github.com/quic-go/quic-go v0.57.1
	github.com/quic-go/webtransport-go v0.9.1-0.20251115050751-b7714a748e1a
	github.com/stretchr/testify v1.11.1
)

// Removed replace directives to maintain upstream canonical modules. Use upstream canonical modules instead.
replace github.com/quic-go/quic-go => github.com/okdaichi/quic-go v0.0.0-20251206231436-9d4d0b2820c2

replace github.com/quic-go/webtransport-go => github.com/okdaichi/webtransport-go v0.9.2-0.20251206232458-65e9257355b8

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dunglas/httpsfv v1.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/quic-go/qpack v0.6.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	golang.org/x/crypto v0.41.0 // indirect
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
