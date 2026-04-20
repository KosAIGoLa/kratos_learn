package biz

import "github.com/google/wire"

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewEmailUsecase,
	NewSMTPSender,
	wire.Bind(new(EmailSender), new(*SMTPSender)),
)
