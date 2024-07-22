from flask import Flask, request, jsonify

app = Flask(__name__)

@app.route('/', methods=['POST'])
def process_prompt():
    data = request.get_json()
    print(data)
    o = str(data)
    return o, 200

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=3101, debug=True)

