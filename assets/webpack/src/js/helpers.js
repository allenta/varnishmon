import Toast from 'bootstrap/js/dist/toast';

/******************************************************************************
 * DATES.
 ******************************************************************************/

/**
 * Converts a date to a Unix timestamp in seconds.
 *
 * @param {Date} date - The date to convert.
 * @returns {number} The Unix timestamp in seconds.
 */
export function dateToUnix(date) {
  return Math.round(date.getTime() / 1000);
}

/**
 * Converts a Unix timestamp in seconds to a date.
 *
 * @param {number} unix - The Unix timestamp to convert.
 * @returns {Date} The date instance.
 */
export function unixToDate(unix) {
  return new Date(unix * 1000);
}

/******************************************************************************
 * NOTIFICATIONS.
 ******************************************************************************/

/**
 * Displays a notification with the specified level and message.
 *
 * @param {string} level - The notification level: 'info', 'success', 'warning',
 * or 'error'.
 * @param {string} message - The notification message.
 */
export function notify(level, message) {
  // Search for required DOM elements: container and template.
  const container = document.getElementById('notifications');
  const template = document.getElementById('notification-template');

  // Create the notification and append it to the container.
  const notification = template.content.cloneNode(true).firstElementChild;
  const btnClose = notification.querySelector('.btn-close');
  let decoration = '';
  if (level === 'info') {
    notification.classList.add('bg-primary', 'text-white');
    btnClose.classList.add('btn-close-white');
    decoration = 'fa-circle-info';
  } else if (level === 'success') {
    notification.classList.add('bg-success', 'text-white');
    btnClose.classList.add('btn-close-white');
    decoration = 'fa-circle-check';
  } else if (level === 'warning') {
    notification.classList.add('bg-warning', 'text-body');
    decoration = 'fa-circle-exclamation';
  } else if (level === 'error') {
    notification.classList.add('bg-danger', 'text-white');
    btnClose.classList.add('btn-close-white');
    decoration = 'fa-circle-exclamation';
  } else {
    notification.classList.add('bg-secondary', 'text-white');
    btnClose.classList.add('btn-close-white');
    decoration = 'fa-comment';
  }
  notification.querySelector('.toast-body').innerHTML = `<i class="fa-solid ${decoration}"></i> ` + message;
  container.appendChild(notification);

  // Initialize notification as a Bootstrap toast and show it.
  const toast = new Toast(notification, {
    autohide: true,
    delay: 10000,
  });
  toast.show();

  // Listen for the hidden event to remove the notification from the DOM.
  notification.addEventListener('hidden.bs.toast', () => {
    notification.remove();
  });
}

/**
 * Sets up DOM observers to intercept Plotly notifications and display them
 * instead using varnishmon own notification system.
 */
export function overridePlotlyNotificationsSystem() {
  // Observe the body to detect when the plotly-notifier container is added.
  const bodyObserver = new MutationObserver((mutationsList, _observer) => {
    for (const mutation of mutationsList) {
      if (mutation.type === 'childList') {
        mutation.addedNodes.forEach(node => {
          if (node.classList && node.classList.contains('plotly-notifier')) {
            // Hide plotly-notifier container.
            node.style.display = 'none';

            // Stop observing for the plotly-notifier container. It will
            // remain in the DOM for now on.
            bodyObserver.disconnect();

            // Send any initial messages as notifications.
            node.querySelectorAll('.notifier-note').forEach(note => {
              const message = note.querySelector('span').innerText;
              notify('info', message);
            });

            // Observe the plotly-notifier container for new messages.
            const observer = new MutationObserver((mutationsList, _observer) => {
              for (const mutation of mutationsList) {
                if (mutation.type === 'childList') {
                  mutation.addedNodes.forEach(node => {
                    if (node.classList && node.classList.contains('notifier-note')) {
                      const message = node.querySelector('span').innerText;
                      notify('info', message);
                    }
                  });
                }
              }
            });
            observer.observe(node, { childList: true });
          }
        });
      }
    }
  });
  bodyObserver.observe(document.body, { childList: true });
}
