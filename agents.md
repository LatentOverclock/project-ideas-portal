If the agents-md submodule (git) is not initialized yet. This has to be done as a first step:  
`git submodule update --init agents-md`. 

Project-specific requirement changes should be added as versioned delta files (`requirements/project.vN.md`).
For delta interpretation, read `agents-md/implementation.md` → **Project Requirements / Versions**.

The organization-wide implementation constraints in this file must be followed strictly:
[agents-md/implementation.md](agents-md/implementation.md)
