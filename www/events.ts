interface CompletionEvent {
  id: string;
  reminder_id: string;
  completed_by: string;
  completed_at: string;
}

interface Family {
  id: string;
  name: string;
  members: string[];
}

interface Reminder {
  id: string;
  title: string;
  description: string;
  due_date: string;
  completed: boolean;
  completed_at?: string;
  family_id: string;
  family_member: string;
  recurrence?: {
    type: string;
    days?: string[];
    date?: number;
    end_date?: string;
  };
}

function loadMenubar() {
   $("#menu-placeholder").load("components/menu.html", function() {
    // Highlight current page in menu
    const currentPage = window.location.pathname.split('/').pop();
    $(`.nav-link[href="${currentPage}"]`).addClass('active');
  });
}

$(document).ready(function() {
  loadMenubar();

  const loadingState = $('#loading-state');
  const memberSelector = $('#member-selector');
  const timeRange = $('#time-range');
  const eventsContainer = $('#events-container');
  const eventsList = $('.events-list');

  // Parse query parameters
  const urlParams = new URLSearchParams(window.location.search);
  const familyIdParam = urlParams.get('familyId');
  const memberNameParam = urlParams.get('member');

  // Update URL with selected member
  function updateURL(familyId: string, memberName: string) {
    const url = new URL(window.location.href);
    url.searchParams.set('familyId', familyId);
    url.searchParams.set('member', memberName);
    window.history.pushState({}, '', url);
  }

  // Load family members for selector
  function loadFamilyMembers() {
    loadingState.removeClass('d-none');
    memberSelector.hide();

    $.get('/families', function(families: Family[]) {
      memberSelector.empty().append('<option value="">Choose your name...</option>');
      
      families.forEach(family => {
        const optgroup = $('<optgroup>').attr('label', family.name);
        family.members.forEach(member => {
          const value = `${family.id}:${member}`;
          const option = $('<option>')
            .val(value)
            .text(member);
          
          // Select the member from query parameters
          if (family.id === familyIdParam && member === memberNameParam) {
            option.prop('selected', true);
            loadEvents(family.id, member);
          }
          
          optgroup.append(option);
        });
        memberSelector.append(optgroup);
      });

      loadingState.addClass('d-none');
      memberSelector.show();
    });
  }

  function formatDate(dateStr: string): string {
    const date = new Date(dateStr);
    return date.toLocaleString();
  }

  function isToday(date: Date): boolean {
    const today = new Date();
    return date.getDate() === today.getDate() &&
           date.getMonth() === today.getMonth() &&
           date.getFullYear() === today.getFullYear();
  }

  function isThisWeek(date: Date): boolean {
    const now = new Date();
    const sevenDaysAgo = new Date(now);
    sevenDaysAgo.setDate(now.getDate() - 7);
    
    return date >= sevenDaysAgo && date <= now;
  }

  function isThisMonth(date: Date): boolean {
    const now = new Date();
    const thirtyDaysAgo = new Date(now);
    thirtyDaysAgo.setDate(now.getDate() - 30);
    
    return date >= thirtyDaysAgo && date <= now;
  }

  function filterEvents(events: CompletionEvent[], reminderMap: Map<string, Reminder>): CompletionEvent[] {
    return events.filter(event => {
      const date = new Date(event.completed_at);
      const range = timeRange.val();

      switch(range) {
        case 'today':
          return isToday(date);
        case 'week':
          return isThisWeek(date);
        case 'month':
          return isThisMonth(date);
        default:
          return true;
      }
    });
  }

  async function loadEvents(familyId: string, memberName: string) {
    loadingState.removeClass('d-none');
    eventsList.addClass('d-none');

    try {
      // First load all reminders to get their details
      const reminders: Reminder[] = await $.get('/reminders');
      const reminderMap = new Map<string, Reminder>();
      
      // Filter reminders for the selected family and member
      const memberReminders = reminders.filter(r => 
        r.family_id === familyId && r.family_member === memberName
      );
      
      memberReminders.forEach(r => reminderMap.set(r.id, r));

      // Load completion events for each reminder
      const allEvents: CompletionEvent[] = [];
      
      for (const reminder of memberReminders) {
        try {
          const events: CompletionEvent[] = await $.get(`/reminders/${reminder.id}/completion-events`);
          allEvents.push(...events);
        } catch (error) {
          // If a reminder has no completion events, the endpoint might return 404
          // This is normal, so we continue to the next reminder
          console.debug(`No completion events found for reminder ${reminder.id}`);
        }
      }

      const filteredEvents = filterEvents(allEvents, reminderMap);
      displayEvents(filteredEvents, reminderMap);
    } catch (error) {
      console.error('Failed to load events:', error);
      eventsContainer.html(`
        <div class="alert alert-danger">
          Failed to load completion events. Please try again.
        </div>
      `);
    } finally {
      loadingState.addClass('d-none');
      eventsList.removeClass('d-none');
    }
  }

  function displayEvents(events: CompletionEvent[], reminderMap: Map<string, Reminder>) {
    eventsContainer.empty();

    if (events.length === 0) {
      eventsContainer.html(`
        <div class="empty-state">
          <p>No completion events found for the selected time range.</p>
        </div>
      `);
      return;
    }

    events.sort((a, b) => new Date(b.completed_at).getTime() - new Date(a.completed_at).getTime());

    events.forEach(event => {
      const reminder = reminderMap.get(event.reminder_id);
      if (!reminder) return;

      const card = $(`
        <div class="card completion-event mb-3">
          <div class="card-body">
            <h5 class="card-title">${reminder.title}</h5>
            <h6 class="card-subtitle mb-2 text-muted">
              Completed by ${event.completed_by} on ${formatDate(event.completed_at)}
            </h6>
            <p class="card-text">${reminder.description}</p>
          </div>
        </div>
      `);
      eventsContainer.append(card);
    });
  }

  // Event handlers
  memberSelector.on('change', function() {
    const value = $(this).val() as string;
    if (value) {
      const [familyId, memberName] = value.split(':');
      updateURL(familyId, memberName);
      loadEvents(familyId, memberName);
    } else {
      eventsContainer.empty();
    }
  });

  timeRange.on('change', function() {
    const value = memberSelector.val() as string;
    if (value) {
      const [familyId, memberName] = value.split(':');
      loadEvents(familyId, memberName);
    }
  });

  // Initial load
  loadFamilyMembers();
});
