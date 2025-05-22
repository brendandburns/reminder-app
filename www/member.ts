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
  family_id: string;
  family_member: string;
  recurrence?: {
    type: string;
    days?: string[];
    date?: number;
    end_date?: string;
  };
}

interface ReminderView {
  today: Reminder[];
  all: Reminder[];
}

$(document).ready(function() {
  const loadingState = $('#loading-state');
  const memberSelector = $('#member-selector');

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

  // Load all families to populate the member selector
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
            loadMemberReminders(family.id, member);
            $('#member-content').show();
          }
          
          optgroup.append(option);
        });
        memberSelector.append(optgroup);
      });

      loadingState.addClass('d-none');
      memberSelector.show();
    });
  }

  // Check if a reminder is due today
  function isDueToday(reminder: Reminder): boolean {
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    
    // Check direct due date
    const dueDate = new Date(reminder.due_date);
    if (dueDate.getFullYear() === today.getFullYear() &&
        dueDate.getMonth() === today.getMonth() &&
        dueDate.getDate() === today.getDate()) {
      return true;
    }

    // Check recurrence pattern
    if (reminder.recurrence) {
      switch (reminder.recurrence.type) {
        case 'weekly':
          const weekday = today.toLocaleDateString('en-US', { weekday: 'long' }).toLowerCase();
          return reminder.recurrence.days?.includes(weekday) || false;
        
        case 'monthly':
          return reminder.recurrence.date === today.getDate();
        
        default:
          return false;
      }
    }

    return false;
  }

  // Add this helper at the top or near your other helpers
  function isSameDay(date1: Date, date2: Date): boolean {
    return (
      date1.getFullYear() === date2.getFullYear() &&
      date1.getMonth() === date2.getMonth() &&
      date1.getDate() === date2.getDate()
    );
  }

  // Load and filter reminders
  function loadMemberReminders(familyId: string, memberName: string) {
    $('#member-reminders').html(`
      <div class="text-center">
        <div class="spinner-border text-primary" role="status">
          <span class="visually-hidden">Loading reminders...</span>
        </div>
      </div>
    `);

    $.get('/reminders', function(reminders: Reminder[]) {
      const memberReminders = reminders.filter(r => 
        r.family_id === familyId && 
        r.family_member === memberName
      );

      const now = new Date();

      // Incomplete reminders: not completed, and for recurring, not completed today
      const incompleteReminders = memberReminders.filter(r => {
        // @ts-ignore
        if (r.recurrence && r.recurrence.type !== "once" && r.completed_at) {
          const completedAt = new Date(r.completed_at);
          if (isSameDay(now, completedAt)) {
            return false;
          }
        }
        return !r.completed;
      });

      // Completed reminders due today:
      // - Non-recurring: completed === true and due today
      // - Recurring: completed_at is today and due today
      const completedReminders = memberReminders.filter(r => {
        const dueToday = isDueToday(r);
        if (!dueToday) return false;
        // @ts-ignore
        if (r.recurrence && r.recurrence.type !== "once" && r.completed_at) {
          const completedAt = new Date(r.completed_at);
          return isSameDay(now, completedAt);
        }
        return r.completed;
      });

      const viewType = $('input[name="view-type"]:checked').val() as string;

      // For the 'all' view, always display recurring reminders, even if completed today
      let finalReminders: Reminder[];
      if (viewType === 'completed') {
        displayReminders(completedReminders, 'completed');
        return;
      } else if (viewType === 'today') {
        finalReminders = incompleteReminders.filter(isDueToday);
      } else if (viewType === 'all') {
        // Show all incomplete reminders, but always include recurring reminders
        const recurringReminders = memberReminders.filter(r => r.recurrence && r.recurrence.type !== 'once');
        // Use a Set to avoid duplicates
        const allSet = new Set(incompleteReminders.map(r => r.id));
        recurringReminders.forEach(r => allSet.add(r.id));
        finalReminders = memberReminders.filter(r => allSet.has(r.id));
      } else {
        finalReminders = incompleteReminders;
      }
      displayReminders(finalReminders, viewType);
    });
  }

  // Display reminders in the UI
  function displayReminders(reminders: Reminder[], viewType: string) {
    const list = $('#member-reminders').empty();
      
    if (reminders.length === 0) {
      list.append(`
        <div class="text-center text-muted p-3">
          No reminders ${viewType === 'today' ? 'due today' : 'found'}
        </div>
      `);
      return;
    }

    reminders.forEach(r => {
      const date = new Date(r.due_date);
      const formattedDate = new Intl.DateTimeFormat('default', {
        dateStyle: 'full',
        timeStyle: 'short'
      }).format(date);

      let recurrenceText = '';
      if (r.recurrence && r.recurrence.type !== 'once') {
        if (r.recurrence.type === 'weekly') {
          recurrenceText = `Weekly on ${r.recurrence.days?.join(', ')}`;
        } else if (r.recurrence.type === 'monthly') {
          recurrenceText = `Monthly on day ${r.recurrence.date}`;
        }
        if (r.recurrence.end_date) {
          recurrenceText += ` until ${new Date(r.recurrence.end_date).toLocaleDateString()}`;
        }
      }

      const isDue = isDueToday(r);
      
      list.append(`
        <div class="card mb-3 ${isDue ? 'border-warning' : ''}">
          <div class="card-body">
            <div class="d-flex justify-content-between align-items-start">
              <h5 class="card-title">${r.title}</h5>
              ${isDue ? '<span class="badge bg-warning text-dark">Due Today</span>' : ''}
            </div>
            <p class="card-text">${r.description}</p>
            <p class="card-text">
              <small class="text-muted">Due: ${formattedDate}</small>
              ${recurrenceText ? `<br><small class="text-info">🔄 ${recurrenceText}</small>` : ''}
            </p>
            <button class="btn btn-sm btn-success mark-complete" data-id="${r.id}">
              Mark Complete
            </button>
          </div>
        </div>
      `);
    });
  }

  // Handle view type changes
  $('input[name="view-type"]').on('change', function() {
    const [familyId, memberName] = memberSelector.val()?.toString().split(':') || [];
    if (familyId && memberName) {
      loadMemberReminders(familyId, memberName);
    }
  });

  // Handle member selection
  memberSelector.on('change', function() {
    const value = $(this).val() as string;
    if (value) {
      const [familyId, memberName] = value.split(':');
      updateURL(familyId, memberName);
      $('#member-content').show();
      loadMemberReminders(familyId, memberName);
    } else {
      $('#member-content').hide();
      // Clear URL parameters
      window.history.pushState({}, '', window.location.pathname);
    }
  });

  // Handle marking reminders as complete
  $(document).on('click', '.mark-complete', function() {
    const reminderId = $(this).data('id');
    const button = $(this);
    button.prop('disabled', true).html(`
      <span class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span>
      Updating...
    `);

    // Send PATCH with completed: true and completed_at: now
    const now = new Date().toISOString();
    $.ajax({
      url: `/reminders/${reminderId}`,
      method: 'PATCH',
      contentType: 'application/json',
      data: JSON.stringify({ completed: true, completed_at: now }),
      success: function() {
        const [familyId, memberName] = memberSelector.val()?.toString().split(':') || [];
        if (familyId && memberName) {
          loadMemberReminders(familyId, memberName);
        }
      },
      error: function() {
        button.prop('disabled', false).text('Mark Complete');
        alert('Failed to update reminder');
      }
    });
  });
  // Initial load
  loadFamilyMembers();
});