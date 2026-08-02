package cdk8s

type cronImpl struct {
	expression string
}

func (c *cronImpl) ExpressionString() *string {
	result := c.expression
	return &result
}

func NewCron(cronOptions *CronOptions) Cron {
	return newCron(cronOptions)
}

func NewCron_Override(cron Cron, cronOptions *CronOptions) {
	if cron == nil {
		panic("parameter cron is required, but nil was provided")
	}
	implementation := newCron(cronOptions)
	if target, ok := cron.(*cronImpl); ok {
		*target = *implementation
		return
	}
	if !setEmbeddedImplementation(cron, implementation) {
		panic("cdk8s: Cron override must embed cdk8s.Cron")
	}
}

// Create a cron schedule which runs first day of January every year.
func Cron_Annually() Cron {
	return Cron_Schedule(&CronOptions{
		Minute:  cronString("0"),
		Hour:    cronString("0"),
		Day:     cronString("1"),
		Month:   cronString("1"),
		WeekDay: cronString("*"),
	})
}

// Create a cron schedule which runs every day at midnight.
func Cron_Daily() Cron {
	return Cron_Schedule(&CronOptions{
		Minute:  cronString("0"),
		Hour:    cronString("0"),
		Day:     cronString("*"),
		Month:   cronString("*"),
		WeekDay: cronString("*"),
	})
}

// Create a cron schedule which runs every minute.
func Cron_EveryMinute() Cron {
	return Cron_Schedule(&CronOptions{
		Minute:  cronString("*"),
		Hour:    cronString("*"),
		Day:     cronString("*"),
		Month:   cronString("*"),
		WeekDay: cronString("*"),
	})
}

// Create a cron schedule which runs every hour.
func Cron_Hourly() Cron {
	return Cron_Schedule(&CronOptions{
		Minute:  cronString("0"),
		Hour:    cronString("*"),
		Day:     cronString("*"),
		Month:   cronString("*"),
		WeekDay: cronString("*"),
	})
}

// Create a cron schedule which runs first day of every month.
func Cron_Monthly() Cron {
	return Cron_Schedule(&CronOptions{
		Minute:  cronString("0"),
		Hour:    cronString("0"),
		Day:     cronString("1"),
		Month:   cronString("*"),
		WeekDay: cronString("*"),
	})
}

// Create a custom cron schedule from a set of cron fields.
func Cron_Schedule(options *CronOptions) Cron {
	if options == nil {
		panic("parameter options is required, but nil was provided")
	}
	return newCron(options)
}

// Create a cron schedule which runs every week on Sunday.
func Cron_Weekly() Cron {
	return Cron_Schedule(&CronOptions{
		Minute:  cronString("0"),
		Hour:    cronString("0"),
		Day:     cronString("*"),
		Month:   cronString("*"),
		WeekDay: cronString("0"),
	})
}

func newCron(options *CronOptions) *cronImpl {
	if options == nil {
		options = &CronOptions{}
	}
	return &cronImpl{expression: cronValue(options.Minute) + " " +
		cronValue(options.Hour) + " " +
		cronValue(options.Day) + " " +
		cronValue(options.Month) + " " +
		cronValue(options.WeekDay)}
}

func cronValue(value *string) string {
	if value == nil {
		return "*"
	}
	return *value
}

func cronString(value string) *string {
	return &value
}
