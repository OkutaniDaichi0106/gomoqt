package moqtransfork

type SetupRequest struct {
	Path              string
	SupportedVersions []Version
	Parameters        Parameters
}

type SetupResponceWriter interface {
	Accept(Version)
	Reject()
}

type SetupHandler interface {
	HandleSetup(SetupRequest, SetupResponceWriter)
}
