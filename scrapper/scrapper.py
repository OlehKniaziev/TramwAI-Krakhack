import requests
import json
from bs4 import BeautifulSoup

# Function to fetch and parse the webpage
def fetch_and_parse(url):
    # Send HTTP request to the URL
    response = requests.get(url)

    # Check if request was successful (status code 200)
    if response.status_code == 200:
        # Parse the content of the page with BeautifulSoup
        soup = BeautifulSoup(response.content, 'html.parser')
        return soup
    else:
        print(f"Failed to retrieve the page. Status code: {response.status_code}")
        return None

counter = 0
json_data = []
# AFISHA.PL
def parse_event(url):
    global json_data
    global counter
    soup = fetch_and_parse(url)

    description = soup.find(class_='description').find('strong')
    if description is None:
        description = ""
    else:
        description = description.text

    data = {
        "title": soup.find(class_='page-header-title').text,
        "data": soup.find(class_='calendar-container').get('data-start'),
        "description": description,
        "location": soup.find(class_='calendar-container').get('data-location'),
        "category": soup.find_all(class_='list-inline-item text-nowrap')[-1].text,
        "url": url
    }

    json_data.append(data)

    #DEBUG
    counter += 1

    #print(counter, url, title, date, description, location, category)

def find_event_links(soup):
    events_div = soup.find(class_='events')
    link_divs = events_div.find_all('div', {'class': 'eventDescription'})
    for div in link_divs:
        link = div.find('a').get('href')
        parse_event("https://www.afisha.pl" + link)

def parse_pages(soup):
    find_event_links(soup)
    sibling = soup.find(class_='page-numbers current').find_next()
    if sibling.name == "a":
        print('NEWPAGE')
        parse_pages(fetch_and_parse(sibling.get('href')))

# Main function to drive the script
if __name__ == '__main__':
    url = 'https://www.afisha.pl/location/Krak%C3%B3w/'
    soup = fetch_and_parse(url)

    if soup:
        parse_pages(soup)
        with open('data.json', 'w') as f:
            json.dump(json_data, f, indent=4)
        # scrape_data(soup)
