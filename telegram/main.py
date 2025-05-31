from telegram import Update, ReplyKeyboardMarkup
from telegram.ext import (
    Application, CommandHandler, MessageHandler,
    ContextTypes, ConversationHandler, filters
)
import requests

# Conversation states
SLEEP, AWAITING_INPUT, FINDING_EVENTS = range(3)

# Keyboards
main_keyboard = ReplyKeyboardMarkup(
    [["🎭 Znajdź wydarzenia", "💤 Sleep"]],
    resize_keyboard=True
)
cancel_keyboard = ReplyKeyboardMarkup(
    [["❌ Cancel"]],
    resize_keyboard=True
)
back_to_menu_keyboard = ReplyKeyboardMarkup(
    [["⬅️ Wróć do menu"]],
    resize_keyboard=True
)


# Event class
class Event:
    def __init__(self, title, date, description, location, category, link):
        self.name = title
        self.date = date
        self.description = description
        self.location = location
        self.category = category
        self.link = link


# Dummy implementation (real one should fetch from API)
def fetch_events(prompt):
    return [
        Event(
            title="Wielki koncert na Rynku Głównym",
            date="2025-06-15",
            description="Wyjątkowy koncert z udziałem gwiazd.",
            location="Rynek Główny, Kraków",
            category="Muzyka",
            link="https://example.com/koncert"
        )
    ]


# Format an event into HTML message
def format_event(event: Event) -> str:
    return (
        f"🎫 <b>{event.name}</b>\n"
        f"📅 <i>Data:</i> {event.date}\n"
        f"📍 <i>Lokacja:</i> {event.location}\n"
        f"🗂️ <i>Kategoria:</i> <b>{event.category}</b>\n\n"
        f"📝 <i>Opis:</i>\n{event.description}\n\n"
        f"🌐 <i>Link:</i>\n<a href=\"{event.link}\">{event.link}</a>"
    )


# Start command
async def start(update: Update, context: ContextTypes.DEFAULT_TYPE) -> int:
    await update.message.reply_text(
        f"🎉 Cześć {update.effective_user.first_name}! Jestem TramwAI – Twój przewodnik po Krakowie.",
        reply_markup=main_keyboard
    )
    return AWAITING_INPUT


# Handle user input in main menu
async def handle_input(update: Update, context: ContextTypes.DEFAULT_TYPE) -> int:
    text = update.message.text

    if text == "🎭 Znajdź wydarzenia":
        await update.message.reply_text("🔍 Opisz wydarzenie, którego szukasz:", reply_markup=cancel_keyboard)
        return FINDING_EVENTS

    elif text == "💤 Sleep":
        await update.message.reply_text("😴 Bot przeszedł w tryb uśpienia. Kliknij „⬅️ Wróć do menu”, aby wrócić.", reply_markup=back_to_menu_keyboard)
        return SLEEP

    else:
        await update.message.reply_text("❓ Nie rozumiem. Wybierz jedną z opcji z menu.", reply_markup=main_keyboard)
        return AWAITING_INPUT


# Handle sleep mode
async def handle_sleep(update: Update, context: ContextTypes.DEFAULT_TYPE) -> int:
    if update.message.text == "⬅️ Wróć do menu":
        await update.message.reply_text("🌞 Wybudzono z trybu uśpienia!", reply_markup=main_keyboard)
        return AWAITING_INPUT
    else:
        await update.message.reply_text("😴 Kliknij „⬅️ Wróć do menu”, aby kontynuować.", reply_markup=back_to_menu_keyboard)
        return SLEEP


# Handle event search
async def search_events_handle(update: Update, context: ContextTypes.DEFAULT_TYPE) -> int:
    text = update.message.text

    if text == "❌ Cancel":
        await update.message.reply_text("↩️ Wróciłeś do głównego menu.", reply_markup=main_keyboard)
        return AWAITING_INPUT

    events = fetch_events(text)

    if not events:
        await update.message.reply_text("Nie znaleziono żadnych wydarzeń.", reply_markup=cancel_keyboard)
    else:
        for event in events:
            await update.message.reply_text(format_event(event), parse_mode="HTML", reply_markup=cancel_keyboard)

    return FINDING_EVENTS


# Cancel command
async def cancel(update: Update, context: ContextTypes.DEFAULT_TYPE) -> int:
    await update.message.reply_text("❌ Zakończono rozmowę. Do zobaczenia!", reply_markup=main_keyboard)
    return ConversationHandler.END


# Main entry point
def main():
    with open('./token.txt') as f:
        BOT_TOKEN = f.read().strip()

    app = Application.builder().token(BOT_TOKEN).build()

    conv_handler = ConversationHandler(
        entry_points=[CommandHandler("start", start)],
        states={
            AWAITING_INPUT: [MessageHandler(filters.TEXT & ~filters.COMMAND, handle_input)],
            SLEEP: [MessageHandler(filters.TEXT & ~filters.COMMAND, handle_sleep)],
            FINDING_EVENTS: [MessageHandler(filters.TEXT & ~filters.COMMAND, search_events_handle)],
        },
        fallbacks=[CommandHandler("cancel", cancel)],
    )

    app.add_handler(conv_handler)
    app.run_polling()


if __name__ == "__main__":
    main()
