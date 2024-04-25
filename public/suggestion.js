document.addEventListener('DOMContentLoaded', function () {
    const input = document.getElementById('searchRecipient');
    const suggestionsBox = document.getElementById('suggestionsBox');
    let activeSuggestionIndex = -1;

    input.addEventListener('input', function(event) {
        const searchTerm = event.target.value;
        activeSuggestionIndex = -1;
        if (searchTerm.length < 2) {
            suggestionsBox.innerHTML = '';
            suggestionsBox.style.display = 'none';
            return;
        }

        fetch(`/api/suggestions/recipients?term=${encodeURIComponent(searchTerm)}`)
            .then(response => response.json())
            .then(suggestions => {
                if (suggestions.length) {
                    suggestionsBox.innerHTML = suggestions.map((s, index) => 
                        `<div class="suggestion p-2 hover:bg-gray-200 cursor-pointer" data-index="${index}">${s}</div>`
                    ).join('');
                    suggestionsBox.style.display = 'block';
                } else {
                    suggestionsBox.style.display = 'none';
                }
            });
    });

    suggestionsBox.addEventListener('click', function(event) {
        if (event.target.classList.contains('suggestion')) {
            selectSuggestion(event.target.innerText);
        }
    });

    input.addEventListener('keydown', function(event) {
        const suggestions = suggestionsBox.getElementsByClassName('suggestion');
        if (suggestions.length === 0) return;

        if (event.key === 'ArrowDown') {
            event.preventDefault();
            if (activeSuggestionIndex < suggestions.length - 1) activeSuggestionIndex++;
            updateActiveSuggestion(suggestions);
        } else if (event.key === 'ArrowUp') {
            event.preventDefault();
            if (activeSuggestionIndex > 0) activeSuggestionIndex--;
            updateActiveSuggestion(suggestions);
        } else if (event.key === 'Enter' && activeSuggestionIndex >= 0) {
            event.preventDefault();
            selectSuggestion(suggestions[activeSuggestionIndex].innerText);
        }
    });

    function selectSuggestion(value) {
        input.value = value;
        suggestionsBox.innerHTML = '';
        suggestionsBox.style.display = 'none';
    }

    function updateActiveSuggestion(suggestions) {
        Array.from(suggestions).forEach((el, index) => {
            if (index === activeSuggestionIndex) {
                el.classList.add('suggestion-active');
                el.scrollIntoView({ block: 'nearest', behavior: 'smooth' });
            } else {
                el.classList.remove('suggestion-active');
            }
        });
    }

    document.addEventListener('click', function(event) {
        if (!input.contains(event.target) && !suggestionsBox.contains(event.target)) {
            suggestionsBox.style.display = 'none';
        }
    });
});
