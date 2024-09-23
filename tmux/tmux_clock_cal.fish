#!/usr/bin/env fish
function tmux_clock_cal

    while true

        # Select calendars
        set -l calendars "\
            General,\
            Fika Group,\
            Vida classes,\
            raj@itsbrilliant.co,\
            Raj,\
            Company Wide,\
            23BAA480-B581-4A3A-BABA-338C567B2A95, \
            "

        # # Get the next appointment
        set -l next_appointment (\
            icalBuddy \
            --includeCals $calendars \
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

        if test $argv[1] = 1
            if test -z $next_appointment
                set appointment (printf '\uf073  Nothing Planned -')
            else
                # TODO strip emojis, time and trailing whitespace
                set appointment (printf '\uf073  %s -' $next_appointment)
            end
        end

        switch $argv[2]
            case 1
                set time_zone "US/Pacific"
                set city "US Pacific"
            case 2
                set time_zone "US/Central"
                set city "US Central"
            case 3
                set time_zone "US/Eastern"
                set city "US Eastern"
            case 4
                set time_zone "Europe/London"
                set city "London"
            case 5
                set time_zone "Europe/Stockholm"
                set city "Central Europe"
            case 6
                set time_zone "Asia/Shanghai"
                set city "China"
        end

        if test -z $time_zone
            tmux set -g status-right "$appointment #(date +%%H:%%M) "(random)
        else
            tmux set -g status-right "$appointment #(TZ=$time_zone date +%%H:%%M) ($city)"
        end

        sleep 5

    end

end

pkill -f "tmux_clock_cal.fish"

tmux_clock_cal $argv[1] $argv[2]