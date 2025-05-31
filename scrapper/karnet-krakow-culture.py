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

json_data = []

def parse_event(url):
    global json_data
    soup = fetch_and_parse(url)
    try:
        location = soup.find(class_='block-list with-link').find('li').find(class_='label').find('span').text
    except:
        location = ""
    data = {
        "title": soup.find('h1').text,
        "data": soup.find(class_='event-date').find(class_='label').find('span').text,
        "description": soup.find(class_='article-content').text,
        "location": location,
        "category": soup.find('h4').text,
        "url": url
    }
    print(data)
    json_data.append(data)

def find_event_links(url):
    soup = fetch_and_parse(url)
    event_list = soup.find_all(class_='event-item')
    for event in event_list:
        parse_event("https://karnet.krakowculture.pl" + event.find('div').find('a').get('href'))

# Main function to drive the script
if __name__ == '__main__':
    for p in range (1, 55):
        find_event_links(f'https://karnet.krakowculture.pl/wydarzenia?Item_page={p}')
    with open('krakowculturedata.json', 'w') as f:
        json.dump(json_data, f, indent=4)
