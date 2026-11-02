package engine

type Options struct {
	display      Display
	drawer       Drawer
	updater      Updater
	fixedUpdater Updater
	laterUpdater Updater
}

type Option func(*Options)

func DefaultOptions() Options {
	return Options{
		display:      DisplayFunc(noopLayout),
		drawer:       DrawerFunc(noopDraw),
		updater:      UpdaterFunc(noopUpdate),
		fixedUpdater: UpdaterFunc(noopUpdate),
		laterUpdater: UpdaterFunc(noopUpdate),
	}
}

func WithDisplay(display Display) Option {
	return func(o *Options) {
		o.display = display
	}
}

func WithDrawer(drawer Drawer) Option {
	return func(o *Options) {
		o.drawer = drawer
	}
}

func WithUpdater(updater Updater) Option {
	return func(o *Options) {
		o.updater = updater
	}
}

func WithFixedUpdater(fixedUpdater Updater) Option {
	return func(o *Options) {
		o.fixedUpdater = fixedUpdater
	}
}

func WithLaterUpdater(laterUpdater Updater) Option {
	return func(o *Options) {
		o.laterUpdater = laterUpdater
	}
}
