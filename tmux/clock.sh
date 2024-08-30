#!/bin/zsh

# Select calendars
CALENDARS="\
    General,\
    Fika Group,\
    Vida classes,\
    raj@itsbrilliant.co,\
    Raj,\
    Company Wide,\
    23BAA480-B581-4A3A-BABA-338C567B2A95, \
    "

# Get the next appointment
NEXT_APPOINTMENT=$(\
    icalBuddy \
    --includeCals $CALENDARS \
    --includeOnlyEventsFromNowOn \
    --limitItems 1 \
    --bullet '' \
    --propertySeparators '| |' \
    --includeEventProps 'datetime,title' \
    --propertyOrder 'datetime,title' \
    --dateFormat '' \
    --noRelativeDates \
    --excludeEndDates \
    --noCalendarNames \
    'eventsToday+1')

# If the event isn't hidden and there is an event to show, format the event
if ! [[ -v HIDE_APPOINTMENT ]] && [ $NEXT_APPOINTMENT ]; then
    APPOINTMENT=$(printf '\uf073  %s –' $NEXT_APPOINTMENT)
fi

# Choose the timezone based on what the environment specifies
if [ $CURRENT_TZ -eq 1 ]; then
    TIME_ZONE="US/Pacific"
    CITY="US Pacific"
elif [ $CURRENT_TZ -eq 2 ]; then
    TIME_ZONE="US/Central"
    CITY="US Central"
elif [ $CURRENT_TZ -eq 3 ]; then
    TIME_ZONE="US/Eastern"
    CITY="US Eastern"
elif [ $CURRENT_TZ -eq 4 ]; then
    TIME_ZONE="Europe/London"
    CITY="London"
elif [ $CURRENT_TZ -eq 5 ]; then
    TIME_ZONE="Europe/Stockholm"
    CITY="Central Europe"
elif [ $CURRENT_TZ -eq 6 ]; then
    TIME_ZONE="Asia/Shanghai"
    CITY="China"
fi

# Show the regional time if requested
if [ $TIME_ZONE ]; then 
    tmux set -g status-right "$APPOINTMENT #(TZ=$TIME_ZONE date +%%H:%%M) ($CITY)"
    sleep 5
fi

# Show the local time
tmux set -g status-right "$APPOINTMENT #(date +%%H:%%M)"
