from pdf_processor import extract_pdf_content
import json
from sorting_topics import extract_topic_content


def process_json_file_to_topics(
    input_file_path: str, output_file_path: str = "extracted_topics.json"
):
    with open(input_file_path, "r") as f:
        data = json.load(f)
    pages = data["pages"]
    # metadata = data["metadata"]
    extracted_data = []
    for page in pages:
        print(f"Processing page {page['page_number']}")

        extracted_topics = extract_topic_content(
            page["text"],
            {
                "page_number": page["page_number"],
                "raw_char_count": page["char_count"],
                "raw_word_count": page["word_count"],
                "raw_sentence_count": page["sentence_count"],
            },
        )
        extracted_data.append(extracted_topics)
    with open(output_file_path, "w") as f:
        json.dump(extracted_data, f)


def main():
    process_json_file_to_topics(
        "./raw_data_from_pdf/10th-science-ch6-control-and-corordination.json",
        "./topic_sorted_data/10th-science-ch6-control-and-corordination.json",
    )


if __name__ == "__main__":
    main()
