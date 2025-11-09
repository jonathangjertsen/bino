var calendarEl = document.getElementById('calendar');
var calendar = new FullCalendar.Calendar(calendarEl, {
    height: 'auto',
    initialView: calendarEl.dataset["initialview"],
    initialDate: calendarEl.dataset["initialdate"],
    locale: 'nb',
    themeSystem: 'bootstrap5',
    headerToolbar: {
        start: 'title',
        center: 'dayGridMonth dayGridWeek listDay',
        end: 'today prev,next',
    },
    eventSources: [
        {
            url: "/calendar/away",
        },
        {
            url: "/calendar/patientevents",
        }
    ]
});
calendar.render();
