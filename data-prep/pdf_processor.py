from typing import List, Dict, Tuple
from PyPDF2 import PdfReader
import os
from unidecode import unidecode
import re
from datetime import datetime


def extract_pdf_content(pdf_path: str) -> Tuple[List[Dict], Dict]:
    """
    Extract text and metadata from a PDF file.

    Args:
        pdf_path (str): Path to the PDF file

    Returns:
        List[Dict]: List of dictionaries containing text and metadata for each page
        Dict: Metadata for the PDF file
    """
    if not os.path.exists(pdf_path):
        raise FileNotFoundError(f"PDF file not found at path: {pdf_path}")

    # Initialize PDF reader
    reader = PdfReader(pdf_path)

    # Get PDF metadata and clean it
    raw_metadata = reader.metadata
    metadata = clean_metadata(raw_metadata)
    
    

    # Extract text and metadata for each page
    pages = []
    total_char_count = 0
    total_word_count = 0
    total_sentence_count = 0
    
    for page_num, page in enumerate(reader.pages, 1):
        text = clean_text(page.extract_text())
        char_count = len(text)
        word_count = len(text.split())
        sentence_count = len(text.split("."))
        
        # Add to totals
        total_char_count += char_count
        total_word_count += word_count
        total_sentence_count += sentence_count
        
        page_dict = {
            "page_number": page_num,
            "text": text,
            "char_count": char_count,
            "word_count": word_count,
            "sentence_count": sentence_count,
        }
        pages.append(page_dict)
    
    # Add document statistics to metadata
    num_pages = len(pages)
    metadata["num_pages"] = num_pages
    metadata["total_char_count"] = total_char_count
    metadata["total_word_count"] = total_word_count
    metadata["total_sentence_count"] = total_sentence_count
    
    # Calculate averages
    if num_pages > 0:
        metadata["avg_chars_per_page"] = round(total_char_count / num_pages, 2)
        metadata["avg_words_per_page"] = round(total_word_count / num_pages, 2)
        metadata["avg_sentences_per_page"] = round(total_sentence_count / num_pages, 2)
    
    # Add file information
    metadata["file_name"] = os.path.basename(pdf_path)
    metadata["file_size_bytes"] = os.path.getsize(pdf_path)
    
    return pages, metadata


def clean_text(text: str) -> str:
    """
    Clean text by removing extra whitespace, newlines, and special characters.

    Args:
        text (str): The text to clean

    Returns:
        str: Cleaned text with normalized whitespace and removed special characters
    """
    if not text:
        return ""

    # Replace various whitespace characters with a single space
    text = (
        text.replace("\n", " ").replace("\r", " ").replace("\t", " ").replace("\f", " ")
    )

    # Convert all Unicode characters to their ASCII equivalents
    text = unidecode(text)

    # Remove special characters and formatting
    # Remove square symbols and other special characters
    text = text.replace("square", "")

    # Remove repeated text patterns (like "Figure 6.1 Figure 6.1 Figure 6.1")
    # This pattern appears in the data where figure captions are repeated
    parts = text.split()
    cleaned_parts = []
    for i, part in enumerate(parts):
        if i > 0 and part == parts[i - 1]:
            continue
        cleaned_parts.append(part)
    text = " ".join(cleaned_parts)

    # Normalize multiple spaces to a single space
    text = " ".join(text.split())

    return text.strip()


def clean_metadata(metadata: Dict) -> Dict:
    """
    Clean and format PDF metadata.

    Args:
        metadata (Dict): Raw metadata from PDF

    Returns:
        Dict: Cleaned metadata with formatted keys and values
    """
    if not metadata:
        return {}
    
    cleaned_metadata = {}
    
    # Map of PDF metadata keys to cleaner names
    key_mapping = {
        "/Author": "author",
        "/CreationDate": "creation_date",
        "/Creator": "creator",
        "/ModDate": "modification_date",
        "/Producer": "producer",
        "/Title": "title",
        "/Subject": "subject",
        "/Keywords": "keywords",
    }
    
    for key, value in metadata.items():
        # Skip None values
        if value is None:
            continue
            
        # Get the clean key name
        clean_key = key_mapping.get(key, key.lstrip("/").lower())
        
        # Handle date fields
        if "date" in clean_key and isinstance(value, str):
            # Try to parse the date
            try:
                # Handle PDF date format (D:YYYYMMDDHHMMSSZ)
                if value.startswith("D:"):
                    # Extract the date part
                    date_str = value[2:16]  # Get YYYYMMDDHHMMSS
                    # Parse the date
                    parsed_date = datetime.strptime(date_str, "%Y%m%d%H%M%S")
                    # Format as ISO string
                    value = parsed_date.isoformat()
                else:
                    # Try to parse as regular date
                    parsed_date = datetime.fromisoformat(value.replace("Z", "+00:00"))
                    value = parsed_date.isoformat()
            except (ValueError, TypeError):
                # If parsing fails, keep the original value
                pass
        
        cleaned_metadata[clean_key] = value
    
    return cleaned_metadata
