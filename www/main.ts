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

$(document).ready(function () {
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
                  <div>
                    <h5 class="mb-1">${r.title}</h5>
                    <p class="mb-1">${r.description}</p>
                    <small class="text-muted">Due: ${formattedDate}</small>
                    <br>
                    ${r.family_member 
                      ? `<small class="text-primary fw-bold">👤 Assigned to: ${r.family_member}</small>` 
                      : '<small class="text-warning">⚠️ No one assigned</small>'
                    }
                  </div>
                  <span class="badge bg-primary rounded-pill">ID: ${r.id}</span>
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

  // Add reminder
  $('#add-reminder-form').on('submit', function (this: HTMLFormElement, e: Event) {
    e.preventDefault();
    const title = $('#reminder-title').val() as string;
    const description = $('#reminder-description').val() as string;
    const localDate = new Date($('#reminder-due-date').val() as string);
    const due_date = localDate.toISOString(); // Converts to ISO 8601 UTC format
    const family_id = $('#reminder-family').val() as string;
    const family_member = $('#reminder-family-member').val() as string;

    $.ajax({
      url: '/reminders',
      method: 'POST',
      contentType: 'application/json',
      data: JSON.stringify({ 
        title, 
        description, 
        due_date,
        family_id,
        family_member
      }),
      success: function() {
        $('#add-reminder-form')[0].reset();
        loadReminders();
        showDialog('Reminder added successfully!');
      },
      error: function(jqXHR: JQueryXHR, textStatus: string, errorThrown: string) {
        showDialog(`Error adding reminder: ${textStatus}`);
      }
    });
  });

  // Modern Bootstrap dialog
  function showDialog(message: string) {
    let modal = $('#copilot-modal');
    if (modal.length === 0) {
      $('body').append(`
        <div class="modal fade" id="copilot-modal" tabindex="-1" aria-labelledby="copilot-modal-label" aria-hidden="true">
          <div class="modal-dialog modal-dialog-centered">
            <div class="modal-content">
              <div class="modal-header">
                <h5 class="modal-title" id="copilot-modal-label">Success</h5>
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
