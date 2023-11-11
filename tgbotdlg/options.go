package tgbotdlg

import "context"

func WithDialogs(dd ...Dialog) optionFunc {
	return func(o *options) {
		o.dialogs = append(o.dialogs, dd...)
	}
}

func WithInterceptors(ii ...Interceptor) optionFunc {
	return func(o *options) {
		o.interceptors = append(o.interceptors, ii...)
	}
}

func WithOffChatUpdateHandler(h func(context.Context, OffChatUpdate) error) optionFunc {
	return func(o *options) {
		o.offChatUpdateHandler = h
	}
}

type options struct {
	dialogs              []Dialog
	interceptors         []Interceptor
	offChatUpdateHandler func(context.Context, OffChatUpdate) error
}

type optionFunc func(o *options)

func buildOpts(ff []optionFunc) options {
	var opt options
	for _, f := range ff {
		f(&opt)
	}
	return opt
}
