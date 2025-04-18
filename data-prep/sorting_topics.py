from typing import Dict, Optional
from langchain_groq import ChatGroq
from pydantic import BaseModel, Field
from langchain_core.output_parsers import PydanticOutputParser
from langchain_core.utils.function_calling import convert_to_openai_function
from langchain.output_parsers import OutputFixingParser
from langchain_core.prompts import ChatPromptTemplate
from dotenv import load_dotenv
import os

# Load environment variables from .env file
load_dotenv()


class TopicContent(BaseModel):
    """Schema for topic and content extraction."""

    topic: str = Field(description="The main topic or subject of the text chunk")
    content: str = Field(
        description="The detailed content or explanation of the topic - should preserve at least 80% of the information from the original text"
    )


def extract_topic_content(
    text_chunk: str, metadata: Optional[Dict[str, str]] = None
) -> Dict[str, str]:
    """
    Extract topic and comprehensive content from a text chunk using Groq via LangChain.

    Args:
        text_chunk (str): A page or section of text from a textbook
        metadata (Optional[Dict[str, str]]): Optional metadata about the text chunk

    Returns:
        Dict[str, str]: Dictionary containing extracted information and metadata
    """
    # Calculate input metrics for validation later
    input_word_count = len(text_chunk.split())

    # System message for the model with explicit length instructions
    system_message = f"""You are an expert at analyzing educational text and extracting key topics and their content.
    
    Your task is to identify the main topic and create a concise yet comprehensive summary that retains 60-80% of the original content.
    
    IMPORTANT: The original text is approximately {input_word_count} words. Your content extraction should be between {int(input_word_count * 0.6)} and {int(input_word_count * 0.8)} words.
    
    Extract and preserve:
    - Core concepts and key definitions
    - Essential examples that illustrate main points
    - Critical relationships and connections
    - Key technical details and supporting evidence
    - All formulas, equations, and technical specifications
    
    Guidelines for summarizing:
    - Focus on the most significant details and remove redundant information
    - Keep only the text information, don't add external knowledge
    - Maintain the original logical flow and structure
    - Preserve all technical terms and specialized vocabulary
    - Keep the most illustrative examples
    - Include all numerical data and formulas
    - Retain context necessary for understanding
    
    Your goal is to create a focused summary that captures the essence while reducing length by 20-40%. The summary should be detailed enough to serve as a comprehensive reference but more concise than the original."""

    # Convert the Pydantic class to OpenAI function format for function calling
    extract_info_function = convert_to_openai_function(TopicContent)

    # Initialize Groq chat model - increase max_tokens to allow for longer responses
    model = ChatGroq(
        model_name=os.getenv("MODEL_NAME"),
        temperature=0.0,
        max_tokens=4000,  # Increased to allow for longer responses
        api_key=os.getenv("GROQ_API_KEY"),
    )

    # Try function calling approach first
    try:
        # Create the chat prompt template with explicit instructions about length
        prompt = ChatPromptTemplate.from_messages(
            [
                ("system", system_message),
                (
                    "human",
                    f"Please analyze the following text ({input_word_count} words) and extract the topic and COMPREHENSIVE content. Your content extraction should be at least {int(input_word_count * 0.8)} words to preserve most of the original information:\n\n{text_chunk}",
                ),
            ]
        )

        # Bind the function to the model
        model_with_functions = model.bind(
            functions=[extract_info_function], function_call={"name": "TopicContent"}
        )

        # Create the chain
        chain = prompt | model_with_functions

        # Run the chain
        response = chain.invoke({"text": text_chunk})

        # Extract the function call results
        if (
            hasattr(response, "additional_kwargs")
            and "function_call" in response.additional_kwargs
        ):
            import json

            function_args = json.loads(
                response.additional_kwargs["function_call"]["arguments"]
            )
            extracted_info = TopicContent(**function_args)
        else:
            # Fallback to content parsing if function call doesn't work
            raise ValueError(
                "Function calling not supported by model, falling back to parser"
            )

    except Exception as e:
        # Fallback to traditional parsing approach
        print(
            f"Function calling failed with error: {e}. Falling back to parser method."
        )

        # Initialize the output parser with error handling
        parser = PydanticOutputParser(pydantic_object=TopicContent)
        robust_parser = OutputFixingParser.from_llm(parser=parser, llm=model)

        # Create the prompt with format instructions and length requirements
        prompt = f"{system_message}\n\nText to analyze ({input_word_count} words):\n{text_chunk}\n\n{parser.get_format_instructions()}\n\nIMPORTANT: Your content extraction should contain AT LEAST {int(input_word_count * 0.8)} words to preserve most of the original information."

        # Get response from the model
        response = model.invoke(prompt)

        # Parse the response with error handling
        extracted_info = robust_parser.parse(response.content)

    # Calculate additional metrics
    content_length = len(extracted_info.content)
    word_count = len(extracted_info.content.split())
    preservation_ratio = word_count / input_word_count if input_word_count > 0 else 0

    # Check if the extracted content meets the length requirements
    if (
        word_count < input_word_count * 0.7
    ):  # Allow some flexibility (70% rather than 80%)
        # If content is too short, try again with even more explicit instructions
        retry_prompt = f"""You MUST extract more comprehensive content from the text. The original text is {input_word_count} words, but your extraction was only {word_count} words ({int(preservation_ratio * 100)}% of original).
        
        Please re-analyze the text and extract AT LEAST {int(input_word_count * 0.8)} words of content to properly preserve the information.
        
        Original text:
        {text_chunk}
        
        Previous extraction (TOO SHORT):
        {extracted_info.content}
        
        Guidelines:
        - Include MORE details from the original text
        - Keep MORE of the examples and explanations
        - Preserve MORE of the technical information and terminology
        - DO NOT SUMMARIZE - you should be extracting and preserving most of the original content
        """

        # Get a new response
        retry_response = model.invoke(retry_prompt)

        # Parse the content from the retry
        # We'll use a simple heuristic to extract the content - everything after the first paragraph
        content_lines = retry_response.content.split("\n\n", 1)
        if len(content_lines) > 1:
            extracted_info.content = content_lines[1]
        else:
            extracted_info.content = retry_response.content

        # Recalculate metrics
        content_length = len(extracted_info.content)
        word_count = len(extracted_info.content.split())
        preservation_ratio = (
            word_count / input_word_count if input_word_count > 0 else 0
        )

    # Return the results with metadata and preservation metrics
    return {
        "topic": extracted_info.topic,
        "content": extracted_info.content,
        "char_count": content_length,
        "word_count": word_count,
        "original_word_count": input_word_count,
        "preservation_ratio": preservation_ratio,
        "metadata": metadata or {},
    }


# Example usage
if __name__ == "__main__":
    sample_text = """
    The Pythagorean theorem states that in a right triangle, the square of the length of the hypotenuse 
    equals the sum of the squares of the lengths of the other two sides. If we denote the sides as a, b, 
    and c, where c is the hypotenuse, then a² + b² = c². This theorem is fundamental in geometry and has 
    numerous applications in mathematics, physics, engineering, and architecture. It was named after the 
    ancient Greek mathematician Pythagoras, although there is evidence that the relationship was known 
    earlier in various cultures including Babylonian, Indian, and Chinese mathematics. The theorem can be 
    proven in many ways, including algebraic proofs, geometric proofs, and even through physical demonstrations. 
    One common proof involves creating squares on each side of the triangle and showing that the area of the 
    square on the hypotenuse equals the sum of the areas of the squares on the other two sides. The theorem's 
    converse is also true: if the squares of two sides of a triangle equal the square of the third side, 
    then the triangle is a right triangle. This property is often used in construction to ensure that corners 
    are square. The 3-4-5 triangle, where 3² + 4² = 5², is a classic example of a right triangle with integer 
    sides, often called a Pythagorean triple. Other examples include 5-12-13 and 8-15-17. The Pythagorean 
    theorem generalizes to higher dimensions and non-Euclidean geometries with appropriate modifications.
    """

    result = extract_topic_content(
        sample_text, {"source": "geometry_textbook", "page": "42"}
    )
    print(f"Topic: {result['topic']}")
    print(
        f"Content preservation: {result['word_count']}/{result['original_word_count']} words ({result['preservation_ratio']:.1%})"
    )
    print(f"Content: {result['content'][:150]}...")
    print(f"Metadata: {result['metadata']}")
