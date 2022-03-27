package people

import "devzat/pkg/user"

const (
	name     = ""
	argsInfo = ""
	info     = ""
)

type Command struct{}

func (c *Command) Name() string {
	return name
}

func (c *Command) ArgsInfo() string {
	return argsInfo
}

func (c *Command) Info() string {
	return info
}

func (c *Command) IsRest() bool {
	return false
}

func (c *Command) IsSecret() bool {
	return false
}

func (c *Command) Fn(_ string, u *user.User) error {
	u.Room.Broadcast("", `
**Hack Club members**  
Zach Latta     - Founder of Hack Club  
Zachary Fogg   - Hack Club Game Designer  
Matthew        - Hack Club HQ  
Caleb Denio, Safin Singh, Eleeza A  
Jubril, Sarthak Mohanty, Anghe,  
Tommy Pujol, Sam Poder, Rishi Kothari,  
Amogh Chaubey, Ella Xu, Hugo Hu,  
Robert Goll, Tanishq Soni, Arash Nur Iman,  
Temi, Aiden Bai, Ivan Bowman, @epic  
Belle See, Fayd, Benjamin Smith  
Matt Gleich, Jason Appah  
_Possibly more people_


**From my school:**  
Kiyan, Riya, Georgie  
Rayed Hamayun, Aarush Kumar


**From Twitter:**  
Ayush Pathak    @ayshptk  
Bereket         @heybereket  
Sanketh         @SankethYS  
Tony Dinh       @tdinh\_me  
Srushti         @srushtiuniverse  
Surjith         @surjithctly  
Arav Narula     @HeyArav  
Krish Nerkar    @krishnerkar\_  
Amrit           @astro_shenava  
Mudrank Gupta   @mudrankgupta  
Harsh           @harshb__

**And many more have joined!**`)
}
