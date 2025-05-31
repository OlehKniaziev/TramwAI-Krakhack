from telegram import (
    Update, ReplyKeyboardMarkup
)
from telegram.ext import (
    Application, CommandHandler, MessageHandler,
    ContextTypes, ConversationHandler, filters
)

import requests

# Stany rozmowy
AWAITING_INPUT, FAVORITE_EVENTS, FINDING_EVENTS = range(3)

# Klawiatury
main_keyboard = [["🎭 Znajdź wydarzenia", "🌟 Wybrane"]]
cancel_keyboard = [["❌ Cancel"]]
favorite_events_keyboard = [["⬅️ Wróć do menu"]]

# Klasa Event
class Event:
    def __init__(self, name, description, date):
        self.name = name
        self.description = description
        self.date = date

# Przykładowe ulubione wydarzenia z jednym eventem
def fetch_favorite_events():
    return [
        Event("Festiwal Muzyczny", "Wielki koncert na Rynku Głównym", "2025-06-15"),
        Event("Wystawa Sztuki", "Nowoczesna sztuka współczesna w muzeum", "2025-07-01"),
    ]

def fetch_events(prompt):
    url = "https://example.com"
    response =  requests.get(url)
    if response == 200:
        events = response.json
    else:
        print(f"Error {response.status_code}: {response.text}")
    return events

# Start rozmowy
async def start(update: Update, context: ContextTypes.DEFAULT_TYPE) -> int:
    reply_markup = ReplyKeyboardMarkup(main_keyboard, resize_keyboard=True)
    await update.message.reply_text(
        f"🎉 Cześć {update.effective_user.first_name}! Jestem TramwAI – Twój przewodnik po Krakowie.",
        reply_markup=reply_markup
    )
    return AWAITING_INPUT

# Obsługa głównego menu i innych opcji
async def handle_input(update: Update, context: ContextTypes.DEFAULT_TYPE) -> int:
    user_input = update.message.text

    if user_input == "🎭 Znajdź wydarzenia":
        return await prompt_for_event_search(update, context)

    elif user_input == "🌟 Wybrane":
        return await favorite_events_handle(update, context)

    elif user_input == "⬅️ Wróć do menu":
        reply_markup = ReplyKeyboardMarkup(main_keyboard, resize_keyboard=True)
        await update.message.reply_text("↩️ Wróciłeś do głównego menu.", reply_markup=reply_markup)
        return AWAITING_INPUT

    else:
        await update.message.reply_text("❓ Nie rozumiem. Wybierz jedną z opcji z menu.")
        return AWAITING_INPUT

# Pokaz ulubionych wydarzeń (wszystkie naraz) с клавиатурой "Wróć do menu"
async def favorite_events_handle(update: Update, context: ContextTypes.DEFAULT_TYPE) -> int:
    events = fetch_favorite_events()

    reply_markup = ReplyKeyboardMarkup(favorite_events_keyboard, resize_keyboard=True)

    if len(events) == 0:
        await update.message.reply_text("Nie masz zapisanych wydarzeń", reply_markup=reply_markup)
        return FAVORITE_EVENTS
    
    text = "\n\n".join(f"{event.name} ({event.date})\n{event.description}" for event in events)

    await update.message.reply_text(text, reply_markup=reply_markup)
    return FAVORITE_EVENTS

# Prośba o wpisanie zapytania do wyszukiwania wydarzeń
async def prompt_for_event_search(update: Update, context: ContextTypes.DEFAULT_TYPE) -> int:
    reply_markup = ReplyKeyboardMarkup(cancel_keyboard, resize_keyboard=True)
    await update.message.reply_text("🔍 Opisz wydarzenie, którego szukasz:", reply_markup=reply_markup)
    return FINDING_EVENTS

# Obsługa wpisanego zapytania wyszukiwania
async def search_events_handle(update: Update, context: ContextTypes.DEFAULT_TYPE) -> int:
    user_input = update.message.text

    # Если пользователь отменяет — возвращаем в главное меню
    if user_input == "❌ Cancel":
        reply_markup = ReplyKeyboardMarkup(main_keyboard, resize_keyboard=True)
        await update.message.reply_text("↩️ Wróciłeś do głównego menu.", reply_markup=reply_markup)
        return AWAITING_INPUT
    
    events = fetch_events(user_input)
    reply_markup = ReplyKeyboardMarkup(cancel_keyboard, resize_keyboard=True)

    if len(events) == 0:
        await update.message.reply_text("Nie znaleziono żadnych wydarzeń pasujących do Twojego zapytania.", reply_markup=reply_markup)
    else:
        for event in events:
            text = f"{event.name} ({event.date})\n{event.description}"
            await update.message.reply_text(text, reply_markup=reply_markup)

    return FINDING_EVENTS

# Cancle
async def cancel(update: Update, context: ContextTypes.DEFAULT_TYPE) -> int:
    await update.message.reply_text("❌ Zakończono rozmowę. Do zobaczenia!")
    return ConversationHandler.END

def main():
    with open('./token.txt', 'r') as file:
        BOT_TOKEN = file.read().strip()

    app = Application.builder().token(BOT_TOKEN).build()

    conv_handler = ConversationHandler(
        entry_points=[CommandHandler("start", start)],
        states={
            AWAITING_INPUT: [MessageHandler(filters.TEXT & ~filters.COMMAND, handle_input)],
            FAVORITE_EVENTS: [MessageHandler(filters.TEXT & ~filters.COMMAND, handle_input)],
            FINDING_EVENTS: [MessageHandler(filters.TEXT & ~filters.COMMAND, search_events_handle)],
        },
        fallbacks=[CommandHandler("cancel", cancel)],
    )

    app.add_handler(conv_handler)
    app.run_polling()

if __name__ == "__main__":
    main()
