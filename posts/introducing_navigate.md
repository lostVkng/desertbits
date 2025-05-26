+++
title = "Introducing Navigate"
date = 2025-05-26
slug = "navigation"
+++

1. [Key Concepts](#key-concepts)
2. [Innovations in the industry](#innovations-in-the-industry)
3. [Introducing Navigate](#introducing-navigate)

Navigation is one of the first challenges that robotics engineers need to overcome. Humans learn from an early age how to navigate the world. Simple learning lessons such as not running into doors, maintaining a buffer distance from a table, and how to backtrack your steps to a location you were at previously are not known to robots. For robots to operate successfully in the physical world they need to make sense of the environment and plan their movements.

# Key Concepts

Building a navigation path requires knowledge of the current environment and the destination (goal) that you wish to reach. A map is a representation of the world around the robot. Some robots operate only in known environments where the robot has a record of everything in its world. Other robots operate in unknown environments where they must learn about the world around them while operating. Similar to how humans observe new details each time they're in a location.

A common technique in robotics is Simultaneous Localization and Mapping (SLAM) which builds a map of the current environment and estimates the current location using sensor data. SLAM is a chicken and egg problem, to understand your location you need a map but to build a reliable map you need to know your current location. SLAM implementations vary based on the sensors available. Vehicles with LIDAR sensors will use different techniques than vehicles that rely only on cameras.

Once you understand your current position, the next step is planning an optimal path to the goal. Path planners utilize a variety of algorithms to build a route for the robot to follow. 

1. Search based algorithms
2. Sampling based algorithms

Search based algorithms (Djikstra, A*) operate on a graph where the relative distance (cost) is known from each square.  Scalability can be a challenge because these algorithms require defined possible nodes. Sampling based algorithms (RRT) explore space randomly which enables much quicker path creation, at the cost of the path potentially not being the most optimal.

Once the optimal route has been identified, planners will often smooth the path using curves. This is so make the path easier to traverse, a jagged path with a lot of turns is likely slower and creates stress on the robot.

Bringing all these concepts together, for a robot to reach its destination the robotic engineer must first understand its current environment. After planning a route, the robot needs to observe and understand its environment along the way to successfully navigate to its goal.

# Innovations in the industry

When Oussama Khatib released his 1985 [paper](https://khatib.stanford.edu/publications/pdfs/Khatib_1985.pdf) on Artificial Potential Fields, it became foundational literature for decades of robotics research and application. New concepts such as object detection were born. 

Progress is made constantly, but the big leaps appear to group together. Looking back at just the last few years. Self driving cars, autonomous drone delivery, and fully robotic shipping facilities are everywhere and expected. Hardware is more accessible today than even just 2 years ago.

The OSS software ecosystem is well timed to catchup to hardware innovation. Most major projects have not received updates in years.

Last Updates from Major Navigation Projects:

1. [ORB_SLAM2](https://github.com/raulmur/ORB_SLAM2) - 9 years
2. [Google Cartographer](https://github.com/cartographer-project/cartographer) - Abandoned
3. [GMapping](https://github.com/OpenSLAM-org/openslam_gmapping) - 14 years

The next leap forward will only accelerate robotic adoption and accessibility.

# Introducing Navigate

I am announcing the development of the [Navigate library](https://crates.io/crates/navigate). Robotics needs a modern library that is flexible and versatile. Not locked into a singular ecosystem or specific to a type of robot or sensor but usable in a variety of projects. 

Today Navigate includes only the dijkstra's algorithm. Over the next few weeks, i'll be porting other industry standard and modern algorithms into Navigate so that it can used for all navigation needs. The goal is not to have it be a specific node in a framework like Dora or ROS but for it to be a core library that can be used across any ecosystem without having to reinvent the wheel.