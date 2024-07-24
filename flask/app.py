from flask import Flask, request, jsonify
def createapp():
    app = Flask(__name__)

    @app.route('/', methods=['POST'])
    def process_prompt():
        data = request.get_json()
        prompt = data['prompt']
        id     = data['id']
        print(f"The <prompt> of {id} is: {prompt}")
        o = str(data)
        return o, 200
    return app

if __name__ == '__main__':
    app1 = createapp()
    app2 = createapp()
    app3 = createapp()
    app4 = createapp()
    app5 = createapp()

    from threading import Thread

    def run_app1():
        app1.run(port=3101)
    def run_app2():
        app2.run(port=3102)
    def run_app3():
        app3.run(port=3103)
    def run_app4():
        app4.run(port=3104)
    def run_app5():
        app5.run(port=3105)

    Thread(target=run_app1).start()
    Thread(target=run_app2).start()
    Thread(target=run_app3).start()
    Thread(target=run_app4).start()
    Thread(target=run_app5).start()
