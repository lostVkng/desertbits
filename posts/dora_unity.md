+++
title = "Dora-rs & Unity Simulations"
date = 2025-05-11
slug = "dora-unity-simulation"
+++

1. [ROS2 Ecosystem](#ros2-ecosystem)
2. [Dora-rs](#dora-rs)
3. [Robotic Simulation](#robotic-simulation)
4. [Communicating between Dora-rs & Unity](#communicating-between-dora-rs-unity)
5. [Conclusion](#conclusion)


The robotics industry has been innovating at an incredible pace. Every day we're moving reducing the gap between Science Fiction and Science Fact. From Nvidia & Disney's Star Wars inspired [BDX droids](https://www.techradar.com/computing/artificial-intelligence/nvidia-google-and-disneys-ai-powered-star-wars-robot-is-absolutely-the-droid-ive-been-looking-for) to humanoid robots like [Figure](https://www.figure.ai/). To further accelerate the pace of innovation, we need better tooling and easier development.

Most of the Open Source frameworks for robotics were created decades ago and were not designed for modern development workflows. Using them with modern tools is cumbersome and the learning curve is steep. This has led the sophisticated robotics companies to build in house middleware stacks resulting in a large gap between private robotics and the FOSS ecosystem. 

This post will explore advances in the open source robotics stack, focusing on middleware and simulation.

# ROS2 Ecosystem

The Robotic Operating System (ROS) ecosystem is the standard middleware for open source robotics. It brings together a wide range of tooling including a package manager, build system, and set of libraries for interacting with both hardware abstractions and low level device drivers. At the core of ROS is the DDS - Data Distribution Service - which provides a pub/sub messaging model for independent nodes or applications to exchange data. In addition to the core ROS middleware, the ecosystem has a number of apps for working with robots such as Gazeebo for simulation and ROS 2 Navigation for path planning.

While great for its time, the software landscape has evolved. There are better build tools, data messaging options, and development workflows. I've found the ROS ecosystem to be clunky and my build chain always ends up looking like a spaghetti mess. I've been wanting a way to more easily bring together nodes/apps that are best suited for the problem at hand. For example allowing a Rust based route planning app to exchange data with a pytorch model for object detection in python then send the data to a low level hardware controller in C or C++. ROS got its node based workflow right, but the build system kills any real attempt at modularity. 

To address this challenge, I first wrote my own middleware framework. I leveraged [ZeroMQ](https://zeromq.org/) for messaging between the nodes and containerized the nodes using [OCI](https://opencontainers.org/) containers to reduce the build complexity. While it was a fun exercise, I still found myself searching for a better solution. Then, last week I came across [Dora-rs](https://dora-rs.ai/) and I believe it shows real promise.

# Dora-rs

[Dora-rs](https://dora-rs.ai/) is a new framework for building modern robotics applications. It's built around a shared memory model that uses [Apache Arrow](https://arrow.apache.org/) for zero-copy messaging between nodes. Early benchmarking shows impressive results, such as the 17x latency improvement over ROS2. 

![Dora-rs Latency](/assets/dora_ros_latency.png)

Building on modern FOSS technologies such as Arrow which has strong FFI language support, allows Dora-rs to seamlessly integrate with a wide range of languages and tools.

# Robotic Simulation

Robotic simulation is a critical tool for robotics developers who need to prototype and train robots. By utilizing a virtual environment, we can simulate the robot's environment and test its behavior in a safe and cost-effective way. But even the best simulations are not a replacement for real-world testing. Simulation is ultimately another tool in a robotics engineer's toolbox.

Benefits of Simulation:

1. Safety
2. Financial Cost 
3. Time Efficiency
4. Difficult to Create Scenarios

Safety is the most important consideration when designing large scale robotic system. We should consider both the individuals testing the system as well as the environment the robot will eventually be operating in. A common use case for robots is to send them to locations that are unsafe for humans to operate such as deap-sea operations or nuclear plant decommissioning. Testing in these environments is dangerous and expensive for humans, not to mention it's difficult to iterate as these environments are not always accessible. Simulation can give us a safe environment for the developers to iterate in, before eventual production deployment. 

Cost is another consideration for robotics development. It takes a massive amount of iteration before a robot is production grade. If the robot is lost or destroyed on each test, it is both time and financially expensive to iterate. It'll likely take time to recreate the robot, potentially source materials, and ship it to the test location. From a pure cost POV, simulation can save a considerable amount of time and money to get the robot 80% of the way to production with minimal cash outlay.

From my point of view, the biggest difference between simulation and reality is that time does not need to be linear in a simulation. You can effectively compress time by having multiple copies of the robot operating in the same or different environments to cover near unlimited scenarios. 

3 days ago, Sequoia Capital released a [video with Nvidia's Director of AI](https://www.youtube.com/watch?v=_2NijXqBESI), Jim Fan, discussing the future of Embodied AI. I recommend watching it as it's a great take on challenges of Robotics vs non-real world technology such as Large Language Models (LLMs). Nvidia uses simulation that executes at 10,000 times real-time to train their robots. They've coined this as the Simulation Principle. The team was able to simulate 10 years worth of training in 2 hours to learn walking. Then the were able to transfer zero shot without any fine-tuning to a real world robot and it was able to walk immediately. This demonstrates the power of simulation for building real world working robots at exponentially less time than it would take in the real world.

Nvidia released the [GROOT blueprint](https://build.nvidia.com/nvidia/isaac-gr00t-synthetic-manipulation) for synthetic data generation in tandem with [Newton](https://developer.nvidia.com/blog/announcing-newton-an-open-source-physics-engine-for-robotics-simulation) their open-source physics engine. 

![Isaac Lab Simulation](/assets/humanoid_training_isaac_lab.png)

Fan further explains that recreating near infinite scenarios is too complex and expensive. He demonstrates that using open source video creation models mixed with in house data allows the team to create scenarios through prompt engineering. These generated videos enable us to further compress time and provide the robot with training data that would have otherwise taken decades to create.


# Communicating between Dora-rs & Unity

Now let's demonstrate a simple example connecting Dora-rs to a simulation. I'll be using Unity for the simulation and rendering, but other engines such as Unreal Engine, Godot, or even a custom engine are also possible. Unity also works flawlessly on MacOS which is a big plus for me when i'm working on a laptop and don't want to switch to my linux machine.

I've downloaded one of the free car assets from the Unity Asset Store and written some quick C# code to control the car. The code accepts keyboard arrow key and spacebar input to handle acceleration, reverse, steering, and braking. I've also included a few public methods to expose car control outside the CarController class. These leverage Coroutines to allow for smooth control of the car. Imagine the robotic operator says drive x distance, since the car can't teleport (yet), it needs to increase acceleration, track the distance, and then apply brakes to slow down. 

![Car Simulation](/assets/unity_sim_1.gif)


I included a simple TCP server that Unity will own/host. I thought it made sense for Unity to own it so that the middleware layer sends commands to the simulation in a similar way that it would send commands to nodes owning the robotic hardware. The key logic here is that the TCP server waits for a client to connect, then it calls the CarController methods based on the commands received.

I chose TCP because its a simple real-time two way communication protocol and it's easily supported in Unity. We can wrap the server in a class and use it as a first class object in Unity. 

My Dora-rs code is split into two nodes, one is a director that sends commands to the bridge node. The bridge then sends the commands via TCP to the Unity server who then calls the CarController methods.

While both these nodes are written in rust, the architecture of Dora allows nodes to not have to be only Rust/C++ but they can be language agnostic. 

Here is the dataflow yaml that defines the dora-rs nodes.
```yaml
nodes:
  - id: drive_director
    build: cargo build -p drive_director
    path: target/debug/drive_director
    inputs:
      tick: dora/timer/millis/100
    outputs:
      - command
  - id: unity_bridge
    build: cargo build -p unity_bridge
    path: target/debug/unity_bridge
    inputs:
      command: drive_director/command
```

I'll submit a pull request to add the TCP example to the dora node-hub for easy access.

![Car Simulation with Dora](/assets/unity_sim_2.gif)


# Conclusion

The open source robotics ecosystem is releasing new tools that will enable more people to create innovative robots. Dora-rs is a modern take on node based middleware for the modern robotics engineer. Advancements in zero-shot prompting and simulation enable robots to learn at a faster rate than ever before. The modern robotics stack will be built for speed, agnostic of language and tools, and will enable more people to create robots.