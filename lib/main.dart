import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

void main() {
  runApp(
    const ProviderScope(child: MyApp()),
  );
}

class MyApp extends StatelessWidget {
  const MyApp({Key? key}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'App Template',
      theme: ThemeData(
        primarySwatch: Colors.blue,
      ),
      home: const FirstRoute(title: 'App Template'),
    );
  }
}

final counterProvider = StateProvider((ref) => 0);

class FirstRoute extends ConsumerWidget {
  final String title;

  const FirstRoute({
    Key? key,
    required this.title,
  }) : super(key: key);

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return Scaffold(
      appBar: AppBar(
        actions: <Widget>[
          IconButton(
            onPressed: () {
              Navigator.push(
                context,
                MaterialPageRoute(
                  builder: (context) => const SecondRoute(
                    title: 'Second Route',
                  ),
                ),
              );
            },
            icon: const Icon(Icons.contact_page),
          )
        ],
        title: Text(title),
      ),
      body: SafeArea(
        child: Center(
          child: Consumer(
            builder: (BuildContext context, WidgetRef ref, Widget? _) {
              final count = ref.watch(counterProvider).state;
              return Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: <Widget>[
                  const Text(
                    'You have pushed the button this many times:',
                  ),
                  Text(
                    '$count',
                    style: Theme.of(context).textTheme.headline4,
                  ),
                ],
              );
            },
          ),
        ),
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: () => ref.read(counterProvider).state++,
        tooltip: 'Increment',
        child: const Icon(Icons.add),
      ),
    );
  }
}

class SecondRoute extends ConsumerWidget {
  final String title;

  const SecondRoute({
    Key? key,
    required this.title,
  }) : super(key: key);

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return Scaffold(
      appBar: AppBar(
        actions: <Widget>[
          IconButton(
            onPressed: () {
              Navigator.pop(context);
            },
            icon: const Icon(Icons.home),
          )
        ],
        title: Text(title),
      ),
      body: SafeArea(
        child: Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: const <Widget>[
              Text(
                'You are at the second route.',
              ),
              Text(
                'Bonus text',
              ),
            ],
          ),
        ),
      ),
    );
  }
}
