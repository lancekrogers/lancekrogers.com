import { debug, debugError } from '../utils/debug.js';

// Calendar state
let currentWeekOffset = 0;
let availableSlots = [];
let selectedSlot = null;

async function loadSlots() {
  try {
    const response = await fetch("/api/slots");
    availableSlots = await response.json();
    displaySlots();
  } catch (error) {
    debugError("Error loading slots:", error);
    const slotsElement = document.getElementById("time-slots");
    if (slotsElement) {
      slotsElement.innerHTML = '<p class="error">Error loading available times. Please try again.</p>';
    }
  }
}

function displaySlots() {
  const container = document.getElementById("time-slots");
  if (!container) return;
  
  const startDate = new Date();
  startDate.setDate(startDate.getDate() + currentWeekOffset * 7);

  const endDate = new Date(startDate);
  endDate.setDate(endDate.getDate() + 7);

  // Filter slots for current week
  const weekSlots = availableSlots.filter((slot) => {
    const slotDate = new Date(slot.date);
    return slotDate >= startDate && slotDate < endDate;
  });

  if (weekSlots.length === 0) {
    container.innerHTML = '<p class="no-slots">No available times this week. Try another week.</p>';
    return;
  }

  // Group by date
  const slotsByDate = {};
  weekSlots.forEach((slot) => {
    if (!slotsByDate[slot.date]) {
      slotsByDate[slot.date] = [];
    }
    slotsByDate[slot.date].push(slot);
  });

  // Display
  container.innerHTML = "";
  Object.entries(slotsByDate).forEach(([date, slots]) => {
    const dateDiv = document.createElement("div");
    dateDiv.className = "date-group";

    const dateObj = new Date(date + "T00:00:00");
    const dayName = dateObj.toLocaleDateString("en-US", { weekday: "short" });
    const monthDay = dateObj.toLocaleDateString("en-US", { month: "short", day: "numeric" });

    dateDiv.innerHTML = `
      <div class="date-header">
        <div class="day-name">${dayName}</div>
        <div class="month-day">${monthDay}</div>
      </div>
      <div class="time-slots">
        ${slots.map(slot => `
          <button class="time-slot" data-slot-id="${slot.id}" data-date="${slot.date}" data-time="${slot.time}">
            ${slot.time}
          </button>
        `).join("")}
      </div>
    `;

    container.appendChild(dateDiv);
  });

  // Add click handlers
  document.querySelectorAll(".time-slot").forEach((button) => {
    button.addEventListener("click", selectSlot);
  });

  // Update week display
  updateWeekDisplay(startDate, endDate);
}

function updateWeekDisplay(start, end) {
  const display = document.getElementById("current-week");
  if (!display) return;
  
  const startStr = start.toLocaleDateString("en-US", { month: "short", day: "numeric" });
  const endStr = end.toLocaleDateString("en-US", { month: "short", day: "numeric" });
  display.textContent = `${startStr} - ${endStr}`;
}

function selectSlot(e) {
  const button = e.target;
  selectedSlot = {
    id: button.dataset.slotId,
    date: button.dataset.date,
    time: button.dataset.time,
  };

  // Update UI
  document.querySelectorAll(".time-slot").forEach((btn) => btn.classList.remove("selected"));
  button.classList.add("selected");

  // Show booking form
  showBookingForm();
}

function showBookingForm() {
  const calendarContainer = document.getElementById("calendar-container");
  const bookingForm = document.getElementById("booking-form");
  
  if (calendarContainer && bookingForm) {
    calendarContainer.classList.add("hidden");
    bookingForm.classList.remove("hidden");

    // Display selected time
    const dateObj = new Date(selectedSlot.date + "T00:00:00");
    const dateStr = dateObj.toLocaleDateString("en-US", {
      weekday: "long",
      month: "long",
      day: "numeric",
    });
    
    const selectedTimeElement = document.getElementById("selected-time");
    const selectedSlotElement = document.getElementById("selected-slot");
    
    if (selectedTimeElement) {
      selectedTimeElement.textContent = `${dateStr} at ${selectedSlot.time} MT`;
    }
    if (selectedSlotElement) {
      selectedSlotElement.value = selectedSlot.id;
    }
  }
}

function showCalendar() {
  const calendarContainer = document.getElementById("calendar-container");
  const bookingForm = document.getElementById("booking-form");
  
  if (calendarContainer && bookingForm) {
    bookingForm.classList.add("hidden");
    calendarContainer.classList.remove("hidden");
  }
}

export function initializeCalendar() {
  debug('Initializing calendar functionality');
  
  // Add event handlers for calendar navigation
  const prevWeekBtn = document.getElementById("prev-week");
  const nextWeekBtn = document.getElementById("next-week");
  const cancelBookingBtn = document.getElementById("cancel-booking");
  const bookingForm = document.getElementById("booking-details");

  if (prevWeekBtn) {
    prevWeekBtn.addEventListener("click", () => {
      currentWeekOffset--;
      displaySlots();
    });
  }

  if (nextWeekBtn) {
    nextWeekBtn.addEventListener("click", () => {
      currentWeekOffset++;
      displaySlots();
    });
  }

  if (cancelBookingBtn) {
    cancelBookingBtn.addEventListener("click", showCalendar);
  }

  if (bookingForm) {
    bookingForm.addEventListener("submit", async (e) => {
      e.preventDefault();

      const formData = new FormData(e.target);
      const data = Object.fromEntries(formData);

      try {
        const response = await fetch("/api/book", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(data),
        });

        const result = await response.json();

        if (response.ok) {
          // Show confirmation
          const bookingFormEl = document.getElementById("booking-form");
          const confirmationEl = document.getElementById("confirmation");
          
          if (bookingFormEl && confirmationEl) {
            bookingFormEl.classList.add("hidden");
            confirmationEl.classList.remove("hidden");

            const dateObj = new Date(selectedSlot.date + "T00:00:00");
            const dateStr = dateObj.toLocaleDateString("en-US", {
              weekday: "long",
              month: "long",
              day: "numeric",
            });
            
            const confirmationDetails = document.getElementById("confirmation-details");
            if (confirmationDetails) {
              confirmationDetails.textContent = `${dateStr} at ${selectedSlot.time} MT`;
            }
          }
        } else {
          const bookingResponse = document.getElementById("booking-response");
          if (bookingResponse) {
            bookingResponse.innerHTML = `<div class="alert error">${result.message || "Booking failed. Please try again."}</div>`;
          }
        }
      } catch (error) {
        debugError("Booking error:", error);
        const bookingResponse = document.getElementById("booking-response");
        if (bookingResponse) {
          bookingResponse.innerHTML = '<div class="alert error">An error occurred. Please try again.</div>';
        }
      }
    });
  }

  // Load slots
  loadSlots();
}