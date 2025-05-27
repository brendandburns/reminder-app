import { MDCRipple } from 'material-components-web';
import $ from 'jquery';

// Add type imports
interface JQuery {
  modal(action: string): void;
}

interface JQueryXHR {
  responseText: string;
  status: number;
  statusText: string;
}

function loadMenubar() {
   $("#menu-placeholder").load("components/menu.html", function() {
    // Highlight current page in menu
    const currentPage = window.location.pathname.split('/').pop();
    $(`.nav-link[href="${currentPage}"]`).addClass('active');
  });
}

$(document).ready(function () {
  loadMenubar();
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
  }

  // Load families
  function loadFamilies() {
    $.get('/families', function (families: Family[]) {
      // Update family list
      const list = $('#family-list').empty();
      families.forEach((f: Family) => {
        list.append(`
          <li class="list-group-item">
            <h5>${f.name}</h5>
            <small class="text-muted">ID: ${f.id}</small>
            <ul class="list-unstyled ms-3">
              ${f.members.map((m: string) => `<li>${m}</li>`).join('')}
            </ul>
          </li>
        `);
      });

      // Update family dropdown in reminder form
      const familySelect = $('#reminder-family').empty();
      familySelect.append('<option value="">Select a family</option>');
      families.forEach((f: Family) => {
        familySelect.append(`<option value="${f.id}" data-members='${JSON.stringify(f.members)}'>${f.name}</option>`);
      });
    });
  }

  // Load reminders
  function loadReminders() {
    $.ajax({
      url: '/reminders',
      method: 'GET',
      success: function(reminders: Reminder | Reminder[]) {
        console.log('Reminders:', reminders);
        console.log('Type of reminders:', typeof reminders);
        const list = $('#reminder-list').empty();
        if (!reminders || (!Array.isArray(reminders) && typeof reminders !== 'object')) {
          list.append('<li class="list-group-item text-danger">No reminders found.</li>');
          return;
        }

        const reminderArray = Array.isArray(reminders) ? reminders : [reminders];
        if (reminderArray.length === 0) {
          list.append('<li class="list-group-item text-muted">No reminders yet. Add your first reminder!</li>');
          return;
        }

        reminderArray.forEach((r: Reminder) => {
          try {
            const date = new Date(r.due_date);
            const formattedDate = new Intl.DateTimeFormat('default', {
              dateStyle: 'full',
              timeStyle: 'short'
            }).format(date);

            console.log('Reminder:', r); // Debug log
            console.log('Family member:', r.family_member); // Debug log

            list.append(`
              <li class="list-group-item">
                <div class="d-flex justify-content-between align-items-center">
                  <div class="flex-grow-1">
                    <h5 class="mb-1">${r.title}</h5>
                    <p class="mb-1">${r.description}</p>
                    <small class="text-muted">Due: ${formattedDate}</small>
                    <br>
                    ${r.family_member 
                      ? `<small class="text-primary fw-bold">👤 Assigned to: ${r.family_member}</small>` 
                      : '<small class="text-warning">⚠️ No one assigned</small>'
                    }
                  </div>
                  <div class="d-flex align-items-center gap-2">
                    <span class="badge bg-primary rounded-pill">ID: ${r.id}</span>
                    <button class="btn btn-sm btn-outline-danger delete-reminder" data-reminder-id="${r.id}" title="Delete reminder">
                      🗑️
                    </button>
                  </div>
                </div>
              </li>
            `);
          } catch (e) {
            console.error('Error formatting reminder:', e, r);
            list.append(`
              <li class="list-group-item text-danger">
                Error displaying reminder: ${r.title || 'Unknown'}
              </li>
            `);
          }
        });
      },
      error: function(jqXHR: JQueryXHR, textStatus: string, errorThrown: string) {
        $('#reminder-list')
          .empty()
          .append(`<li class="list-group-item text-danger">Failed to load reminders: ${textStatus}</li>`);
        console.error('Failed to load reminders:', textStatus, errorThrown);
      }
    });
  }

  // Add family
  $('#add-family-form').on('submit', function (e: Event) {
    e.preventDefault();
    const name = $('#family-name').val();
    const members = $('#family-members').val()?.toString().split(',').map((m: string) => m.trim());
    $.ajax({
      url: '/families',
      method: 'POST',
      contentType: 'application/json',
      data: JSON.stringify({ name, members }),
      success: function() {
        loadFamilies();
        showDialog('Family added successfully!');
      }
    });
  });

  // Handle family selection change
  $('#reminder-family').on('change', function(this: HTMLSelectElement) {
    const selected = $(this).find(':selected');
    const memberSelect = $('#reminder-family-member').empty();
    memberSelect.append('<option value="">Select a family member</option>');
    
    if (selected.val()) {
      const members = selected.data('members');
      members.forEach((member: string) => {
        memberSelect.append(`<option value="${member}">${member}</option>`);
      });
    }
  });

  // Populate monthly date options
  const monthlyDate = document.getElementById('monthly-date') as HTMLSelectElement;
  for (let i = 1; i <= 31; i++) {
    const option = document.createElement('option');
    option.value = i.toString();
    option.textContent = i.toString();
    monthlyDate.appendChild(option);
  }

  // Handle recurrence type changes
  const recurrenceType = document.getElementById('reminder-recurrence-type') as HTMLSelectElement;
  const weeklyOptions = document.getElementById('weekly-options') as HTMLDivElement;
  const monthlyOptions = document.getElementById('monthly-options') as HTMLDivElement;
  const endDateContainer = document.getElementById('end-date-container') as HTMLDivElement;
  const dueDateContainer = document.getElementById('due-date-container') as HTMLDivElement;
  const dueDateInput = document.getElementById('reminder-due-date') as HTMLInputElement;

  recurrenceType.addEventListener('change', () => {
    weeklyOptions.style.display = 'none';
    monthlyOptions.style.display = 'none';
    endDateContainer.style.display = 'none';

    // Show/hide due date based on recurrence type
    if (recurrenceType.value === 'once') {
      dueDateContainer.style.display = 'block';
      dueDateInput.required = true;
    } else {
      dueDateContainer.style.display = 'none';
      dueDateInput.required = false;
      dueDateInput.value = ''; // Clear the value when hidden
    }

    switch (recurrenceType.value) {
      case 'daily':
        endDateContainer.style.display = 'block';
        break;
      case 'weekly':
        weeklyOptions.style.display = 'block';
        endDateContainer.style.display = 'block';
        break;
      case 'monthly':
        monthlyOptions.style.display = 'block';
        endDateContainer.style.display = 'block';
        break;
    }
  });

  // Initialize the form state
  recurrenceType.dispatchEvent(new Event('change'));

  // Update form submission to include recurrence data
  const addReminderForm = document.getElementById('add-reminder-form') as HTMLFormElement;
  addReminderForm.addEventListener('submit', async (e) => {
    e.preventDefault();

    const title = (document.getElementById('reminder-title') as HTMLInputElement).value;
    const description = (document.getElementById('reminder-description') as HTMLInputElement).value;
    const familyId = (document.getElementById('reminder-family') as HTMLSelectElement).value;
    const familyMember = (document.getElementById('reminder-family-member') as HTMLSelectElement).value;
    const dueDateValue = (document.getElementById('reminder-due-date') as HTMLInputElement).value;
    const endDate = (document.getElementById('reminder-end-date') as HTMLInputElement).value;

    const recurrence: any = {
      type: recurrenceType.value
    };

    if (recurrence.type === 'weekly') {
      const selectedDays = Array.from(document.querySelectorAll('.weekday:checked'))
        .map(cb => (cb as HTMLInputElement).value);
      if (selectedDays.length === 0) {
        alert('Please select at least one day for weekly recurrence');
        return;
      }
      recurrence.days = selectedDays;
    } else if (recurrence.type === 'monthly') {
      recurrence.date = parseInt(monthlyDate.value);
    }

    if (endDate && (recurrence.type === 'daily' || recurrence.type === 'weekly' || recurrence.type === 'monthly')) {
      recurrence.end_date = new Date(endDate).toISOString();
    }

    const reminderData: any = {
      title,
      description,
      family_id: familyId,
      family_member: familyMember,
      recurrence
    };

    // Only include due_date for one-time reminders
    if (recurrence.type === 'once' && dueDateValue) {
      reminderData.due_date = new Date(dueDateValue).toISOString();
    }

    try {
      const response = await fetch('/reminders', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(reminderData),
      });

      if (!response.ok) {
        throw new Error('Failed to create reminder');
      }

      // Reset form and refresh reminder list
      addReminderForm.reset();
      loadReminders();
    } catch (error) {
      console.error('Error creating reminder:', error);
      alert('Failed to create reminder');
    }
  });

  // Delete reminder functionality
  $(document).on('click', '.delete-reminder', function(e) {
    e.preventDefault();
    const reminderId = $(this).data('reminder-id');
    const reminderTitle = $(this).closest('.list-group-item').find('h5').text();
    
    showDeleteConfirmation(reminderId, reminderTitle);
  });

  // Show delete confirmation modal
  function showDeleteConfirmation(reminderId: string, reminderTitle: string) {
    let modal = $('#delete-confirmation-modal');
    if (modal.length === 0) {
      $('body').append(`
        <div class="modal fade" id="delete-confirmation-modal" tabindex="-1" aria-labelledby="delete-confirmation-label" aria-hidden="true">
          <div class="modal-dialog modal-dialog-centered">
            <div class="modal-content">
              <div class="modal-header">
                <h5 class="modal-title" id="delete-confirmation-label">Confirm Delete</h5>
                <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
              </div>
              <div class="modal-body">
                <p>Are you sure you want to delete this reminder?</p>
                <div class="alert alert-warning">
                  <strong id="reminder-title-to-delete"></strong>
                </div>
                <p class="text-muted">This action cannot be undone.</p>
              </div>
              <div class="modal-footer">
                <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Cancel</button>
                <button type="button" class="btn btn-danger" id="confirm-delete-btn">Delete</button>
              </div>
            </div>
          </div>
        </div>
      `);
      modal = $('#delete-confirmation-modal');
    }
    
    // Update the modal with the specific reminder details
    modal.find('#reminder-title-to-delete').text(reminderTitle);
    modal.find('#confirm-delete-btn').off('click').on('click', function() {
      performDelete(reminderId);
      // @ts-ignore
      const bsModal = bootstrap.Modal.getInstance(modal[0]);
      bsModal.hide();
    });
    
    // @ts-ignore
    const bsModal = new bootstrap.Modal(modal[0]);
    bsModal.show();
  }

  // Perform the actual delete operation
  function performDelete(reminderId: string) {
    $.ajax({
      url: `/reminders/${reminderId}`,
      method: 'DELETE',
      success: function() {
        showDialog('Reminder deleted successfully!');
        loadReminders(); // Refresh the reminder list
      },
      error: function(jqXHR: JQueryXHR, textStatus: string, errorThrown: string) {
        console.error('Failed to delete reminder:', textStatus, errorThrown);
        showDialog('Failed to delete reminder: ' + textStatus);
      }
    });
  }

  // Modern Bootstrap dialog
  function showDialog(message: string) {
    let modal = $('#copilot-modal');
    if (modal.length === 0) {
      $('body').append(`
        <div class="modal fade" id="copilot-modal" tabindex="-1" aria-labelledby="copilot-modal-label" aria-hidden="true">
          <div class="modal-dialog modal-dialog-centered">
            <div class="modal-content">
              <div class="modal-header">
                <h5 class="modal-title" id="copilot-modal-label">Notification</h5>
                <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
              </div>
              <div class="modal-body"></div>
              <div class="modal-footer">
                <button type="button" class="btn btn-primary" data-bs-dismiss="modal">OK</button>
              </div>
            </div>
          </div>
        </div>
      `);
      modal = $('#copilot-modal');
    }
    modal.find('.modal-body').text(message);
    // @ts-ignore
    const bsModal = new bootstrap.Modal(modal[0]);
    bsModal.show();
  }

  loadFamilies();
  loadReminders();
});
