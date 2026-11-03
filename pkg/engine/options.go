package engine

type options struct {
	display      Display
	drawer       Drawer
	updater      Updater
	fixedUpdater Updater
	laterUpdater Updater
}

type Option func(*options)

func defaultOptions() options {
	return options{
		display:      DisplayFunc(noopLayout),
		drawer:       DrawerFunc(noopDraw),
		updater:      UpdaterFunc(noopUpdate),
		fixedUpdater: UpdaterFunc(noopUpdate),
		laterUpdater: UpdaterFunc(noopUpdate),
	}
}

func WithDisplay(display Display) Option {
	if display == nil {
		panic("engine: nil display")
	}
	return func(o *options) {
		o.display = display
	}
}

func WithDrawer(drawer Drawer) Option {
	if drawer == nil {
		panic("engine: nil drawer")
	}
	return func(o *options) {
		o.drawer = drawer
	}
}

func WithUpdater(updater Updater) Option {
	if updater == nil {
		panic("engine: nil updater")
	}
	return func(o *options) {
		o.updater = updater
	}
}

func WithFixedUpdater(fixedUpdater Updater) Option {
	if fixedUpdater == nil {
		panic("engine: nil fixedUpdater")
	}
	return func(o *options) {
		o.fixedUpdater = fixedUpdater
	}
}

func WithLaterUpdater(laterUpdater Updater) Option {
	if laterUpdater == nil {
		panic("engine: nil laterUpdater")
	}
	return func(o *options) {
		o.laterUpdater = laterUpdater
	}
}
